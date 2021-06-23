# Fingerprints

Files are identified and tracked using fingerprints. A fingerprint is the first `N` bytes of the file, with the default for `N` being `1000`. 

### Fingerprint Growth

When a file is smaller than `N` bytes, the fingerprint is the length of the file. A fingerprint that is less than `N` bytes will be compared to other fingerprints using a prefix check. As the file grows, its fingerprint will be updated, until it reaches the full size of `N`.

### Deduplication of Files

Multiple files with the same fingerprint are handled as if they are the same file. 

Most commonly, this circumstance is observed during file rotation that depends on a copy/truncate strategy. After copying the file, but before truncating the original, two files with the same content briefly exist. If the `file_input` operator happens to observe both files at the same time, it will detect a duplicate fingerprint and ingest only one of the files.

If logs are replicated to multiple files, or if log files are copied manually, it is not understood to be of any significant value to ingest the duplicates. As a result, fingerprints are not designed to differentiate between these files, and double ingestion of the same content is not supported automatically.

In some rare circumstances, a logger may print a very verbose preamble to each log file. When this occurs, fingerprinting may fail to differentiate files from one another. This can be overcome by customizing the size of the fingerprint using the `fingerprint_size` setting.

# Readers

Readers are a convenience struct, which exist for the purpose of managing files and their associated metadata. 

### Contents

A Reader contains the following:
- File handle (may be open or closed)
- File fingerprint
- File offset (aka checkpoint)
- File path
- Decoder (dedicated instance to avoid concurrency issues)

### Functionality

As implied by the name, Readers are responsible for consuming data as it is written to a file.

Before a Reader begins consuming, it will seek the file's last known offset. Then it will begine reading from there.

While a file is shorter than the length of a fingerprint, its Reader will continuously append to the fingerprint, as it consumes newly written data.

A Reader consumes a file using an `bufio.Scanner`, with the Scanner's buffer size defined by the `max_log_size`setting, and the Scanner's split func defined by the `multiline` setting. 

As each log is read from the file, it is decoded according to the `encoding` function, and then emitted from the operator. 

The Reader's offset is updated accordingly when a log is emitted.


### Persistence

Readers are always instantiated with an open file handle. Eventually, the file handle is closed, but the Reader is not discarded. Rather, it is maintained for a fixed number of "poll cycles" (see Polling section below) as a reference to the file's metadata, which may be useful for detecting files that have been moved or copied, and for recalling metadata such as the file's previous path.

Readers are maintained for a fixed period of time, and then discarded.

When the `file_input` operator makes use of a persistence mechanism to save and recall its state, it is simply Setting and Getting a slice of Readers. These Readers contain all the information necessary to pick up exactly where the operator left off.


# Polling

The file system is polled on a regular interval, defined by the `poll_interval` setting. 

Each poll cycle runs through a series of steps which are presented below.

At a very high level, each poll cycle operates as three phases:
1. Finish work that was started in the previous poll cycle.
2. Begin work that will carry over to the next cycle.
3. Allow some time to pass.

Because state is carried over from one poll cycle to the next, the following detailed description is presented starting mid-cycle, at a point where there is no preexisting state.


### Detailed Poll Cycle

1. Matching
    1. Find files that match the `include` setting. The files are known only by their paths.
    2. Discard any of these files that match the `exclude` setting.
    3. As a special case, on the first poll cycle, a warning is printed if no files are matched. Execution continues regardless.
2. Queueing
    1. If the number of matched files is less than or equal to the maximum degree of concurrency, as defined by the `max_concurrent_files` setting, then no queueing occurs.
    2. Else, queueing occurs, which means the following:
        - Matched files are split into two sets, such that the first is small enough to respect `max_concurrent_files`, and the second contains the remaining files (called the queue).
        - The current poll interval will begin processing the first set of files, just as if they were the only ones found during the matching phase.
        - Subsequent poll cycles will pull matches off of the queue, until the queue is empty.
        - The `max_concurrent_files` setting is respected at all times.
        - TODO: Use a worker pool to process the queue as quickly as possible, rather than waiting for poll cycles to trigger batching.
3. Opening
    1. Each of the matched files is opened. Note:
        - A small amount of time has passed since the file was matched.
        - It is possible that is has been moved or deleted by this point.
        - Only a minimum set of operations should occur between file matching and opening.
        - If an error occurs while opening, it is logged.
4. Fingerprinting
    1. The first `N` bytes of each file are read. (See fingerprinting section above.)
5. Exclusion
    1. Empty files are closed immediately and discarded. (There is nothing to read.)
    2. Fingerprints found in this batch are cross referenced against each other to detect duplicates. Duplicate files are closed immediately and discarded.
        - In the vast majority of cases, this occurs during file rotation that uses the copy/truncate method. (See fingerprinting section above.)
6. Reader Creation
    1. Each file handle is wrapped into a `Reader` along with some metadata. (See Reader section above)
        - During the creation of a `Reader`, the file's fingerprint is crossreferenced with previously known fingerprints.
        - If a file's fingerprint matches one that has recently been seen, then metadata is copied over from the previous iteration of the Reader. Most importantly, the offset is accurately maintained in this way.
        - If a file's fingerprint does not match any recently seen files, then its offset is initialized according to the `start_at` setting.
7. Detection of "Lost" Files
8. Consumption
    1. Lost files are consumed. Since these files are thought to have been rotated to a location where we will not match them again, we will never see them again. 
        - The best we can do is finish consuming their current contents.
        - We can reasonable expect in most cases that these files are no longer being written to.
    2. Matched files (from this poll cycle) are consumed.
        - These file handles will be left open until the end of the next poll cycle.
        - Typically, we can expect to find most of these files again. However, these files are consumed greedily, in case we do not see them again.
    3. All files in both sets are consumed concurrently.
9. Closing
    1. All files from the previous poll cycle are closed.
10. Archiving
    1. Readers created in the current poll cycle are added to the historical record.
11. Pruning
    1. The historical record is purged of Readers that have existsed for 3 generations.
        - This number is somewhat arbitrary, and should probably be made configurable. However, its exact purpose is quite obscure.
12. Persistence
    1. The historical record of readers is synced to whatever persistence mechanism was provided to the operator.
13. End of poll cycle. 
    1. At this point, the operator sits idle until the poll timer fires again.

_Note: This is the actual start of poll cycle._

14. Dequeueing
    1. If any matches are queued from the previous cycle, an appropriate number are dequeued, and processed that same as would a newly matched set of files.
15. Aging
    1. If no queued files were left over from the previous cycle, then all previously matched files have been consumed, and we are ready to query the file system again. Prior to doing so, we will increment the "generation" of all historical Readers.


# Additional Details

### Startup Logic

Whenever the operator starts, it:
- Requests the historical record of Readers, as described in steps 10-12 of the poll cycle.
- Starts the polling timer.

### Shutdown Logic

When the operator shuts down, it:
- Closes any open files.


### Known Limitations

On Windows, rotation of files using the Move/Create strategy may cause errors and loss of data, because Golang does not currently support the Windows mechanism for `FILE_SHARE_DELETE`.
