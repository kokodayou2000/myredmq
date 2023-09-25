package services

var createMagNow *CreateMag

type CreateMag struct{}

func (create CreateMag) CreateTopic(topicName string) int {
	// 调用 接口创建 topic 实际上是 streams 或者成为队列
	return 0
}

type CreateMagInterface interface {
	CreateTopic(topicName string) int
}

func NewCreateMag() {
	createMagNow = &CreateMag{}
}

func GetCreateMag() *CreateMag {
	if createMagNow == nil {
		NewCreateMag()
	}
	return createMagNow
}
