package main

func main() {
	t, p, a := recvTargetInfo()
	client := NewClient(t, p, a)

	// TODO: 是が非でもresponseから自力でデータを取得したい。
	client.connectTCPServer()
}
