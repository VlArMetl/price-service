package pricing

import "fmt"

// Money — денежная сумма в копейках.
//
// Деньги никогда не хранят во float64: двоичная дробь не может точно
// представить 0.1, поэтому 0.1 + 0.2 != 0.3, и на миллионах товаров
// копейки начинают «теряться». Целое число копеек считается точно.
type Money int64

// Rub создаёт сумму из целых рублей: Rub(150) == 150,00 ₽.
func Rub(rubles int64) Money {
	return Money(rubles * 100)
}

// Percent возвращает pct процентов от суммы (с округлением вниз до копейки).
func (m Money) Percent(pct int) Money {
	return m * Money(pct) / 100
}

// RoundToRuble округляет сумму до целых рублей по правилам арифметики:
// 149,49 → 149,00; 149,50 → 150,00.
func (m Money) RoundToRuble() Money {
	return (m + 50) / 100 * 100
}

// String печатает сумму в человекочитаемом виде: «149.90 ₽».
// Благодаря этому методу fmt.Println(price) выводит рубли, а не копейки.
func (m Money) String() string {
	sign := ""
	if m < 0 {
		sign = "-"
		m = -m
	}
	return fmt.Sprintf("%s%d.%02d ₽", sign, m/100, m%100)
}
