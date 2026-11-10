// Точка входа сервиса.
//
// Пока здесь только демонстрация расчёта цены. В следующих шагах main
// будет читать конфиг, подключаться к PostgreSQL, Redis и Kafka
// и запускать gRPC-сервер.
package main

import (
	"fmt"
	"log"

	"github.com/VlArMetl/price-service/internal/pricing"
)

func main() {
	calc := pricing.NewCalculator(pricing.DefaultConfig)

	examples := []struct {
		name string
		in   pricing.Input
	}{
		{"Гречка 900 г", pricing.Input{PurchasePrice: pricing.Rub(80)}},
		{"Кофе в зёрнах 1 кг", pricing.Input{PurchasePrice: pricing.Rub(1000), CompetitorPrice: pricing.Rub(1190)}},
		{"Молоко 1 л (срок истекает завтра)", pricing.Input{PurchasePrice: pricing.Rub(70), Perishable: true, DaysToExpiry: 1}},
	}

	for _, ex := range examples {
		res, err := calc.Calculate(ex.in)
		if err != nil {
			log.Fatalf("%s: %v", ex.name, err)
		}
		fmt.Printf("%-36s закупка %-12v → цена %-12v (%s)\n", ex.name, ex.in.PurchasePrice, res.Price, res.Reason)
	}
}
