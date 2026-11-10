package pricing

import (
	"errors"
	"testing"
)

// Table-driven тесты — стандартный стиль в Go: один цикл и таблица случаев.
// Добавить новый случай = добавить одну строку в таблицу.
func TestCalculate(t *testing.T) {
	tests := []struct {
		name       string
		in         Input
		wantPrice  Money
		wantReason Reason
	}{
		{
			name:       "обычная наценка 30%",
			in:         Input{PurchasePrice: Rub(100)},
			wantPrice:  Rub(130),
			wantReason: ReasonTargetMarkup,
		},
		{
			name:       "конкурент дешевле — подстраиваемся",
			in:         Input{PurchasePrice: Rub(100), CompetitorPrice: Rub(120)},
			wantPrice:  Rub(120),
			wantReason: ReasonCompetitor,
		},
		{
			name:       "конкурент слишком дешёвый — не ниже минимальной наценки",
			in:         Input{PurchasePrice: Rub(100), CompetitorPrice: Rub(105)},
			wantPrice:  Rub(110),
			wantReason: ReasonCompetitor,
		},
		{
			name:       "конкурент дороже — его цена не важна",
			in:         Input{PurchasePrice: Rub(100), CompetitorPrice: Rub(150)},
			wantPrice:  Rub(130),
			wantReason: ReasonTargetMarkup,
		},
		{
			name:       "скоропорт, срок ещё большой — без скидки",
			in:         Input{PurchasePrice: Rub(100), Perishable: true, DaysToExpiry: 5},
			wantPrice:  Rub(130),
			wantReason: ReasonTargetMarkup,
		},
		{
			name:       "скоропорт, 2 дня — скидка 15% и округление",
			in:         Input{PurchasePrice: Rub(100), Perishable: true, DaysToExpiry: 2},
			wantPrice:  Rub(111), // 130 − 15% = 110,50 → 111
			wantReason: ReasonExpiryDiscount,
		},
		{
			name:       "скоропорт, 1 день — скидка 30%",
			in:         Input{PurchasePrice: Rub(100), Perishable: true, DaysToExpiry: 1},
			wantPrice:  Rub(91),
			wantReason: ReasonExpiryDiscount,
		},
		{
			name:       "скоропорт, последний день — скидка 50%",
			in:         Input{PurchasePrice: Rub(100), Perishable: true, DaysToExpiry: 0},
			wantPrice:  Rub(65),
			wantReason: ReasonExpiryDiscount,
		},
		{
			name: "скидка по сроку считается от цены конкурента",
			in: Input{
				PurchasePrice: Rub(100), CompetitorPrice: Rub(120),
				Perishable: true, DaysToExpiry: 1,
			},
			wantPrice:  Rub(84), // 120 − 30%
			wantReason: ReasonExpiryDiscount,
		},
		{
			name:       "рост цены ограничен 20%",
			in:         Input{PurchasePrice: Rub(100), CurrentPrice: Rub(100)},
			wantPrice:  Rub(120),
			wantReason: ReasonChangeLimited,
		},
		{
			name:       "падение цены ограничено 20%",
			in:         Input{PurchasePrice: Rub(100), CurrentPrice: Rub(200)},
			wantPrice:  Rub(160),
			wantReason: ReasonChangeLimited,
		},
		{
			name:       "изменение в пределах 20% — без ограничения",
			in:         Input{PurchasePrice: Rub(100), CurrentPrice: Rub(125)},
			wantPrice:  Rub(130),
			wantReason: ReasonTargetMarkup,
		},
		{
			name: "уценку по сроку ограничение не сдерживает",
			in: Input{
				PurchasePrice: Rub(100), CurrentPrice: Rub(200),
				Perishable: true, DaysToExpiry: 0,
			},
			wantPrice:  Rub(65),
			wantReason: ReasonExpiryDiscount,
		},
		{
			name:       "копейки округляются до рубля",
			in:         Input{PurchasePrice: 9999}, // 99,99 ₽ → 129,98 ₽ → 130 ₽
			wantPrice:  Rub(130),
			wantReason: ReasonTargetMarkup,
		},
		{
			name:       "цена не бывает меньше 1 рубля",
			in:         Input{PurchasePrice: 1}, // 1 копейка
			wantPrice:  Rub(1),
			wantReason: ReasonTargetMarkup,
		},
	}

	calc := NewCalculator(DefaultConfig)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Calculate(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Price != tt.wantPrice {
				t.Errorf("price = %v, want %v", got.Price, tt.wantPrice)
			}
			if got.Reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", got.Reason, tt.wantReason)
			}
		})
	}
}

func TestCalculateInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		in   Input
	}{
		{"нулевая закупочная цена", Input{PurchasePrice: 0}},
		{"отрицательная закупочная цена", Input{PurchasePrice: -100}},
		{"отрицательная текущая цена", Input{PurchasePrice: Rub(100), CurrentPrice: -1}},
		{"отрицательная цена конкурента", Input{PurchasePrice: Rub(100), CompetitorPrice: -1}},
		{"просроченный товар", Input{PurchasePrice: Rub(100), Perishable: true, DaysToExpiry: -1}},
	}

	calc := NewCalculator(DefaultConfig)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calc.Calculate(tt.in)
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("err = %v, want ErrInvalidInput", err)
			}
		})
	}
}
