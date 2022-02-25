module github.com/satimoto/go-lsp

go 1.16

require (
	github.com/joho/godotenv v1.4.0
	github.com/lightningnetwork/lnd v0.14.2-beta.rc2
	github.com/satimoto/go-datastore v0.1.2-0.20220225213440-5f4c0c959a53
	github.com/satimoto/go-datastore-mocks v0.1.2-0.20220225121048-b31df3a69220
	google.golang.org/grpc v1.43.0
)

replace git.schwanenlied.me/yawning/bsaes.git => github.com/Yawning/bsaes v0.0.0-20180720073208-c0276d75487e
