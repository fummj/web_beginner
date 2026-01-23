package main

func main() {
	client := NewClient()
	client.recvTargetInfo()

	// TODO: 是が非でもresponseから自力でデータを取得したい。
	client.connectTCPServer()
}
