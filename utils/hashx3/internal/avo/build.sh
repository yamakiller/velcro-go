go run sse.go avx.go gen.go -avx2 -out ./avx2.s
go run sse.go avx.go gen.go -sse2 -out ./sse2.s