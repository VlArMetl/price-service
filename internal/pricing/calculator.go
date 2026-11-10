// Package pricing считает базовую цену товара.
//
// Это «ядро» сервиса: чистая бизнес-логика без базы данных, сети и Kafka.
// Такой код легко тестировать: на вход данные, на выходе цена.
// Всё остальное (PostgreSQL, gRPC, Redis, Kafka) появится в следующих шагах
// и будет лишь доставлять сюда данные и забирать результат.
package pricing

import (
	"errors"
	"fmt"
)

// ErrInvalidInput возвращается, если входные данные не имеют смысла.
// Вызывающий код может проверить его через errors.Is(err, ErrInvalidInput).
var ErrInvalidInput = errors.New("invalid pricing input")

// Config — настройки правил ценообразования.
type Config struct {
	TargetMarkup int // целевая наценка к закупочной цене, %
	MinMarkup    int // минимальная наценка: ниже не опускаемся ради конкурента, %
	MaxChange    int // максимальное изменение цены за один пересчёт, %
}

// DefaultConfig — разумные значения по умолчанию.
var DefaultConfig = Config{
	TargetMarkup: 30,
	MinMarkup:    10,
	MaxChange:    20,
}

// Input — всё, что известно о товаре в момент расчёта.
type Input struct {
	PurchasePrice   Money // закупочная цена, обязательна и > 0
	CurrentPrice    Money // текущая цена на витрине; 0 — товар ещё не продавался
	CompetitorPrice Money // цена у конкурента; 0 — неизвестна
	Perishable      bool  // скоропортящийся товар (молоко, хлеб, зелень)
	DaysToExpiry    int   // сколько дней осталось до конца срока годности
}

// Reason объясняет, какое правило определило итоговую цену.
// Без этого на вопрос «почему молоко подорожало?» не ответить.
type Reason string

const (
	ReasonTargetMarkup   Reason = "target_markup"    // обычная наценка
	ReasonCompetitor     Reason = "competitor_match" // подстроились под конкурента
	ReasonExpiryDiscount Reason = "expiry_discount"  // скидка из-за срока годности
	ReasonChangeLimited  Reason = "change_limited"   // упёрлись в ограничение скачка цены
)

// Result — итог расчёта.
type Result struct {
	Price  Money
	Reason Reason
}

// Calculator считает цену по заданным правилам.
type Calculator struct {
	cfg Config
}

// NewCalculator создаёт калькулятор с заданными настройками.
func NewCalculator(cfg Config) *Calculator {
	return &Calculator{cfg: cfg}
}

// Calculate рассчитывает базовую цену товара.
//
// Правила применяются по порядку:
//  1. Целевая наценка к закупочной цене.
//  2. Если конкурент дешевле — подстраиваемся под него, но не ниже минимальной наценки.
//  3. Скоропортящийся товар с истекающим сроком уценяем: лучше продать дешевле, чем списать.
//  4. Цена не может измениться больше чем на MaxChange% за раз (кроме уценки по сроку):
//     резкие скачки подрывают доверие покупателей.
//  5. Округляем до целых рублей.
func (c *Calculator) Calculate(in Input) (Result, error) {
	if err := validate(in); err != nil {
		return Result{}, err
	}

	// 1. Целевая наценка.
	price := in.PurchasePrice + in.PurchasePrice.Percent(c.cfg.TargetMarkup)
	reason := ReasonTargetMarkup

	// 2. Конкурент.
	floor := in.PurchasePrice + in.PurchasePrice.Percent(c.cfg.MinMarkup)
	if in.CompetitorPrice > 0 && in.CompetitorPrice < price {
		price = max(in.CompetitorPrice, floor)
		reason = ReasonCompetitor
	}

	// 3. Срок годности.
	if in.Perishable {
		if discount := expiryDiscount(in.DaysToExpiry); discount > 0 {
			price -= price.Percent(discount)
			reason = ReasonExpiryDiscount
		}
	}

	// 4. Ограничение скачка цены.
	if in.CurrentPrice > 0 && reason != ReasonExpiryDiscount {
		upper := in.CurrentPrice + in.CurrentPrice.Percent(c.cfg.MaxChange)
		lower := in.CurrentPrice - in.CurrentPrice.Percent(c.cfg.MaxChange)
		if price > upper {
			price, reason = upper, ReasonChangeLimited
		} else if price < lower {
			price, reason = lower, ReasonChangeLimited
		}
	}

	// 5. Округление; цена не может стать нулевой.
	price = max(price.RoundToRuble(), Rub(1))

	return Result{Price: price, Reason: reason}, nil
}

// expiryDiscount возвращает скидку в процентах в зависимости от остатка срока годности.
func expiryDiscount(daysToExpiry int) int {
	switch {
	case daysToExpiry <= 0:
		return 50 // последний день
	case daysToExpiry == 1:
		return 30
	case daysToExpiry == 2:
		return 15
	default:
		return 0
	}
}

func validate(in Input) error {
	switch {
	case in.PurchasePrice <= 0:
		return fmt.Errorf("%w: purchase price must be positive, got %d", ErrInvalidInput, in.PurchasePrice)
	case in.CurrentPrice < 0:
		return fmt.Errorf("%w: current price must not be negative, got %d", ErrInvalidInput, in.CurrentPrice)
	case in.CompetitorPrice < 0:
		return fmt.Errorf("%w: competitor price must not be negative, got %d", ErrInvalidInput, in.CompetitorPrice)
	case in.Perishable && in.DaysToExpiry < 0:
		return fmt.Errorf("%w: product is already expired", ErrInvalidInput)
	}
	return nil
}
