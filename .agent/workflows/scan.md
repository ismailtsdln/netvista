# NetVista v2.0 Scanning Workflow

How to perform an advanced visual reconnaissance scan with NetVista v2.0.

To perform a complete scan with the new Hexagonal Architecture:

1. Build the latest binary:

// turbo
```bash
go build -o netvista cmd/netvista/main.go
```

2. Run a scan with standard options:

```bash
echo "example.com" | ./netvista scan -o [output_dir] -p 80,443 -c 10 -t 30s
```

3. View the results:

```bash
./netvista serve -d [output_dir]
```

4. For incremental scans (skipping processed targets), run the same command again on the same output directory.
