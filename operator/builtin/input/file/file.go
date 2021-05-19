// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v3"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
)

// InputOperator is an operator that monitors files for entries
type InputOperator struct {
	helper.InputOperator

	Include            []string
	Exclude            []string
	FilePathField      entry.Field
	FileNameField      entry.Field
	PollInterval       time.Duration
	SplitFunc          bufio.SplitFunc
	MaxLogSize         int
	MaxConcurrentFiles int

	persister operator.Persister

	knownFiles     []*Reader
	queuedMatches  []string
	readerQueue    []*Reader
	tailingReaders map[string][]*Reader

	startAtBeginning bool

	fingerprintSize int

	encoding helper.Encoding

	wg         sync.WaitGroup
	firstCheck bool
	cancel     context.CancelFunc
}

// Start will start the file monitoring process
func (f *InputOperator) Start(persister operator.Persister) error {
	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel
	f.firstCheck = true

	f.persister = persister
	// Load offsets from disk
	if err := f.loadLastPollFiles(ctx); err != nil {
		return fmt.Errorf("read known files from database: %s", err)
	}

	// Start polling goroutine
	f.startPoller(ctx)

	return nil
}

// Stop will stop the file monitoring process
func (f *InputOperator) Stop() error {
	f.cancel()
	f.wg.Wait()
	f.knownFiles = nil
	f.cancel = nil
	for _, readers := range f.tailingReaders {
		for _, reader := range readers {
			f.closeFile(reader.file, reader.Path)
		}
	}
	f.tailingReaders = nil
	return nil
}

// startPoller kicks off a goroutine that will poll the filesystem periodically,
// checking if there are new files or new logs in the watched files
func (f *InputOperator) startPoller(ctx context.Context) {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		globTicker := time.NewTicker(f.PollInterval)
		defer globTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-globTicker.C:
			}

			f.poll(ctx)
		}
	}()
}

const RotatedFileTrackingLimit = 10

// poll checks all the watched paths for new entries
func (f *InputOperator) poll(ctx context.Context) {
	matches := getMatches(f.Include, f.Exclude)
	if f.firstCheck && len(matches) == 0 {
		f.Warnw("no files match the configured include patterns", "include", f.Include)
	}
	for _, path := range matches {
		reader, err := f.isNewFileToTail(path)
		if err != nil {
			continue
		}
		if reader == nil {
			continue
		}

		if _, ok := f.tailingReaders[path]; !ok { // If path doesn't exist => new file to tail
			if f.firstCheck {
				if f.startAtBeginning {
					f.Infow("Started watching file", "path", path)
				} else {
					f.Infow("Started watching file from end. To read preexisting logs, configure the argument 'start_at' to 'beginning'", "path", path)
				}
				if err := reader.InitializeOffset(f.startAtBeginning); err != nil {
					f.Errorw("Failed to initialize offset for "+path, zap.Error(err))
				}
			} else {
				f.Infow("Started watching file", "path", path)
			}
			f.tailingReaders[path] = []*Reader{reader}
		} else { // If path exists
			f.Infow("Log rotation detected. Started watching file", "path", path)
			f.tailingReaders[path] = append(f.tailingReaders[path], reader)
			// Limit tracking the iteration of rotated files
			if len(f.tailingReaders[path]) > RotatedFileTrackingLimit {
				var firstReader *Reader
				firstReader, f.tailingReaders[path] = f.tailingReaders[path][0], f.tailingReaders[path][1:]
				f.closeFile(firstReader.file, firstReader.Path)
			}
		}
	}
	f.firstCheck = false

	readerCount := 0
	for _, readers := range f.tailingReaders {
		readerCount += len(readers)
	}

	if len(f.readerQueue) <= readerCount {
		for _, readers := range f.tailingReaders {
			f.readerQueue = append(f.readerQueue, readers...)
		}
	}

	count := min0(f.MaxConcurrentFiles, readerCount)
	count = min0(count, len(f.readerQueue))
	polledReaders := f.readerQueue[:count]
	f.readerQueue = f.readerQueue[count:]
	var wg sync.WaitGroup
	for _, reader := range polledReaders {
		wg.Add(1)
		go func(r *Reader) {
			defer wg.Done()
			r.ReadToEnd(ctx)
		}(reader)
	}

	// Wait until all the reader goroutines are finished
	wg.Wait()
	f.syncLastPollFiles(ctx)
}

// getMatches gets a list of paths given an array of glob patterns to include and exclude
func getMatches(includes, excludes []string) []string {
	all := make([]string, 0, len(includes))
	for _, include := range includes {
		matches, _ := filepath.Glob(include) // compile error checked in build
	INCLUDE:
		for _, match := range matches {
			for _, exclude := range excludes {
				if itMatches, _ := doublestar.PathMatch(exclude, match); itMatches {
					continue INCLUDE
				}
			}

			for _, existing := range all {
				if existing == match {
					continue INCLUDE
				}
			}

			all = append(all, match)
		}
	}

	return all
}

// isNewFileToTail compares fingerprints with already tailing files to see if it is a new file or not
func (f *InputOperator) isNewFileToTail(path string) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		f.Errorw("Failed to open file", zap.Error(err))
		return nil, err
	}
	fp, err := f.NewFingerprint(file)
	if err != nil {
		f.Errorw("Failed to make FingerPrint", zap.Error(err))
		f.closeFile(file, path)
		return nil, err
	}
	if len(fp.FirstBytes) == 0 {
		f.closeFile(file, path)
		return nil, nil
	}

	newReader, err := f.NewReader(path, file, fp)
	if err != nil {
		f.Errorw("Failed to make reader for "+path, zap.Error(err))
		f.closeFile(file, path)
		return nil, err
	}

	existingReader, exist := f.findFingerprintMatch(fp)
	if exist {
		if existingReader.Path != path { // chunked rotated file. need to tail this file from previous offset.
			f.Infow("Detected rotated file with a new name. Restoring the correct offset info from the original file.", "path", path)
			newReader.Offset = existingReader.Offset
			f.removeReader(existingReader)
			return newReader, nil
		} else {
			// already tailing
			f.closeFile(file, path)
			return nil, nil
		}
	}
	return newReader, nil
}

func (f *InputOperator) removeReader(reader *Reader) {
	f.closeFile(reader.file, reader.Path)
	for i, r := range f.tailingReaders[reader.Path] {
		if reader.Fingerprint.StartsWith(r.Fingerprint) {
			f.tailingReaders[reader.Path] = append(f.tailingReaders[reader.Path][:i], f.tailingReaders[reader.Path][i+1:]...)
			if len(f.tailingReaders[reader.Path]) == 0 {
				delete(f.tailingReaders, reader.Path)
			}
		}
	}
	for i := 0; i < len(f.readerQueue); i++ {
		r := f.readerQueue[i]
		if reader.Fingerprint.StartsWith(r.Fingerprint) {
			f.readerQueue = append(f.readerQueue[:i], f.readerQueue[i+1:]...)
			i -= 1
		}
	}
}

func (f *InputOperator) findFingerprintMatch(fp *Fingerprint) (*Reader, bool) {
	for _, readers := range f.tailingReaders {
		for _, reader := range readers {
			if fp.StartsWith(reader.Fingerprint) {
				return reader, true
			}
		}
	}
	return nil, false
}

const knownFilesKey = "knownFiles"

// syncLastPollFiles syncs the most recent set of files to the database
func (f *InputOperator) syncLastPollFiles(ctx context.Context) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	// Encode the number of known files.
	if err := enc.Encode(len(f.tailingReaders)); err != nil {
		f.Errorw("Failed to encode known files", zap.Error(err))
		return
	}

	// Encode each known file. For files that has same path, encode only latest one.
	for _, readers := range f.tailingReaders {
		lastReader := readers[len(readers)-1]
		if err := enc.Encode(lastReader); err != nil {
			f.Errorw("Failed to encode known files", zap.Error(err))
		}
	}

	if err := f.persister.Set(ctx, knownFilesKey, buf.Bytes()); err != nil {
		f.Errorw("Failed to sync to database", zap.Error(err))
	}
}

// syncLastPollFiles loads the most recent set of files to the database
func (f *InputOperator) loadLastPollFiles(ctx context.Context) error {
	encoded, err := f.persister.Get(ctx, knownFilesKey)
	if err != nil {
		return err
	}

	f.tailingReaders = map[string][]*Reader{}
	if encoded == nil {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(encoded))

	// Decode the number of entries
	var knownFileCount int
	if err := dec.Decode(&knownFileCount); err != nil {
		return fmt.Errorf("decoding file count: %w", err)
	}
	// Decode each of the known files
	for i := 0; i < knownFileCount; i++ {
		decodedReader, err := f.NewReader("", nil, nil)
		if err != nil {
			return err
		}
		if err = dec.Decode(decodedReader); err != nil {
			return err
		}
		path := decodedReader.Path
		file, err := os.Open(path)
		if err != nil {
			f.Errorw("Failed to open file while recovering checkpoints", zap.Error(err))
			f.tailingReaders[path] = []*Reader{decodedReader}
			continue
		}

		restoredReader, err := f.NewReader(path, file, decodedReader.Fingerprint)
		if err != nil {
			f.Errorw("Failed to restore a Reader", zap.Error(err), "path", path)
			continue
		}
		restoredReader.Offset = decodedReader.Offset
		f.tailingReaders[path] = []*Reader{restoredReader}
	}

	return nil
}

func (f *InputOperator) closeFile(file *os.File, path string) {
	err := file.Close()
	if err != nil {
		f.Warnw("Error closing a file", "file", path, "err", err)
	}
}
