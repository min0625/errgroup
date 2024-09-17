# errgroup
A recoverable errgroup based on `x/sync/errgroup` that can recover from panics. Panics are caught and re-panicked in the Wait function.
