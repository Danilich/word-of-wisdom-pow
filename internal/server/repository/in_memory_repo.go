package repository

import (
	"context"
	"errors"
	"math/rand"
	"time"
	"wisdom-pow/internal/server/domain"
)

// InMemoryRepository implements in-memory storage
type InMemoryRepository struct {
	quotes []domain.Quote
	rand   *rand.Rand
}

func NewInMemoryRepository() *InMemoryRepository {
	source := rand.NewSource(time.Now().UnixNano())
	repository := &InMemoryRepository{
		quotes: []domain.Quote{
			{Text: "Человек создан для счастья, как птица для полета.", Author: "Максим Горький"},
			{Text: "Счастье не в том, чтобы делать всегда, что хочешь, а в том, чтобы всегда хотеть того, что делаешь.", Author: "Лев Толстой"},
			{Text: "Если ты будешь любить жизнь, то и жизнь будет любить тебя.", Author: "Фёдор Достоевский"},
			{Text: "Жизнь дается один раз, и хочется прожить ее бодро, осмысленно, красиво.", Author: "Антон Чехов"},
			{Text: "Надо любить жизнь больше, чем смысл жизни.", Author: "Фёдор Достоевский"},
			{Text: "Быть самим собой в мире, который постоянно пытается сделать вас чем-то другим, — величайшее достижение.", Author: "Ральф Уолдо Эмерсон"},
			{Text: "Никогда не поздно стать тем, кем тебе следовало быть.", Author: "Джордж Элиот"},
			{Text: "Единственный способ делать великую работу — любить то, что вы делаете.", Author: "Стив Джобс"},
			{Text: "Жизнь — это то, что с тобой происходит, пока ты строишь другие планы.", Author: "Джон Леннон"},
			{Text: "Лучше зажечь одну свечу, чем проклинать темноту.", Author: "Конфуций"},
			{Text: "Не бойся, что не получится. Бойся, что не попробуешь.", Author: "Рой Т. Беннетт"},
			{Text: "Самый тяжелый груз — это неисполненный долг.", Author: "Александр Солженицын"},
			{Text: "Кто не идет вперед, тот идет назад. Стоячего положения нет.", Author: "Виссарион Белинский"},
			{Text: "Чтобы поверить в добро, надо начать его делать.", Author: "Лев Толстой"},
			{Text: "Великие умы обсуждают идеи, средние умы обсуждают события, мелкие умы обсуждают людей.", Author: "Элеонора Рузвельт"},
		},
		rand: rand.New(source),
	}

	return repository
}

// GetRandom returns a random quote from the repository
func (r *InMemoryRepository) GetRandom(ctx context.Context) (domain.Quote, error) {
	if ctx.Err() != nil {
		return domain.Quote{}, ctx.Err()
	}

	if len(r.quotes) == 0 {
		return domain.Quote{}, errors.New("no quotes available")
	}

	return r.quotes[r.rand.Intn(len(r.quotes))], nil
}
