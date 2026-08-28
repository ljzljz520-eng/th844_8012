# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	agri-packaging/cmd/packaging-screen	[no test files]
ok  	agri-packaging/internal/audit	0.011s
ok  	agri-packaging/internal/catalog	0.010s
ok  	agri-packaging/internal/dashboard	0.011s
ok  	agri-packaging/internal/grading	0.009s
?   	agri-packaging/internal/model	[no test files]
?   	agri-packaging/internal/planning	[no test files]
?   	agri-packaging/internal/report	[no test files]
ok  	agri-packaging/internal/httpapi	0.009s
ok  	agri-packaging/internal/line	0.011s
--- FAIL: TestPackingRejectsStoppedLine (0.01s)
    service_test.go:28: stopped line accepted progress
FAIL
FAIL	agri-packaging/internal/packing	0.011s
ok  	agri-packaging/internal/shift	0.007s
ok  	agri-packaging/internal/storage	0.010s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/packaging-screen): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/packaging-screen): exit `0`
