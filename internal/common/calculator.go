package common

import (
	"log/slog"
)

// business logic of adapter
// Lazy init - делать мапу переменных в рамках "адаптера".
// В рамках мапы хранятся ссылки на память, где лежат значения переменных
// Сделать одну операцию подсчета, которая будет складывать значения по ссылкам
// И вторая операция - формирование ответа по порядку вызовов print, разыменовывая ссылки

type Calculator struct {
	logger slog.Logger
}

type ExecuteHTTP interface {
	ExecuteHTTP()
}

func Execute() {

}
