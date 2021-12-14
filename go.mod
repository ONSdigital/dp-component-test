module github.com/ONSdigital/dp-component-test

go 1.17

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2

require (
	github.com/ONSdigital/dp-mongodb-in-memory v1.1.0
	github.com/cucumber/godog v0.12.0
	github.com/gorilla/mux v1.8.0
	github.com/kr/text v0.2.0 // indirect
	github.com/maxcnunes/httpfake v1.2.4
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.8.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
