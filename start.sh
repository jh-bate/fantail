#! /bin/bash -eu

. config/env.sh
exec node --trace_gc --trace_gc_verbose --max_new_space_size=16384 --max_old_space_size=48 lib/index.js