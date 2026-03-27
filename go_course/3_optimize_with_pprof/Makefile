generate-easyjson:
	easyjson -all ./structs/user.go

test:
	go test -v

bench:
	go test -bench . -benchmem

get-profiles:
	go test -run=^$$ -bench BenchmarkFast -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1

pprof-cpu:
	go tool pprof 3_optimize_with_pprof.test cpu.out

pprof-mem:
	go tool pprof 3_optimize_with_pprof.test mem.out

