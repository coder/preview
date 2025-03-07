#!/usr/bin/env bash


killbg() {
        echo "Killing background processes..."
        for p in "${pids[@]}" ; do
              echo "Killing $p"
              kill -9 "$p";
        done
}
trap killbg EXIT


SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
echo "Script directory: $SCRIPT_DIR"

pids=()

(cd "$SCRIPT_DIR/../testdata" && pwd && go run ../cmd/preview/main.go web) &
pids+=($!)

(cd "$SCRIPT_DIR/../site" && pnpm install && pnpm dev) &
pids+=($!)

# loop over pids and print
for p in "${pids[@]}" ; do
    echo "Running -> PID: $p"
done

wait $P1 $P2