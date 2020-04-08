package graph


// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"database/sql"
	"time"

	"theodoretliu.com/crossword/server/graph/generated"
	"theodoretliu.com/crossword/server/graph/model"
)

func (r *mutationResolver) UpdateCookie(ctx context.Context, cookie string) (string, error) {
	db := ctx.Value("db").(*sql.DB)

	_, err := db.Exec("UPDATE cookie SET cookie = $1", cookie)

	if err != nil {
		return "", err
	}

	return cookie, nil
}

func (r *queryResolver) User(ctx context.Context, username string) (*model.User, error) {
	db := ctx.Value("db").(*sql.DB)

	row := db.QueryRow("SELECT username, id, name FROM users WHERE username = $1", username)

	var userInfo model.User

	err := row.Scan(&userInfo.Username, &userInfo.ID, &userInfo.Name)

	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	db := ctx.Value("db").(*sql.DB)

	rows, err := db.Query("SELECT id, username, name FROM users")
	if err != nil {
		return nil, err
	}

	res := []*model.User{}

	for rows.Next() {
		var tmp model.User

		rows.Scan(&tmp.ID, &tmp.Username, &tmp.Name)

		res = append(res, &tmp)
	}

	return res, nil
}

func (r *queryResolver) DaysOfTheWeek(ctx context.Context) ([]string, error) {
	days := GetDaysOfTheWeek()

	var stringDays []string

	for _, day := range days {
		stringDays = append(stringDays, day.Format(time.RFC1123Z))
	}

	return stringDays, nil
}

func (r *userResolver) WeeksTimes(ctx context.Context, user *model.User) ([]int, error) {
	return ctx.Value("weeksTimesLoader").(*WeeksTimesLoader).Load(user.ID)
}

func (r *userResolver) WeeklyAverage(ctx context.Context, user *model.User) (int, error) {
	loader := ctx.Value("loader").(*WeeksWorstTimesLoader)
	weeksWorstTimes, err := loader.Load(nil)

	if err != nil {
		return 0, err
	}

	weeksTimes, err := ctx.Value("weeksTimesLoader").(*WeeksTimesLoader).Load(user.ID)
	if err != nil {
		return 0, err
	}

	weights := []int{25, 25, 25, 25, 25, 25, 49}

	total := 0
	totalWeight := 0

	for i := 0; i < 7; i++ {
		if weeksTimes[i] == -1 && weeksWorstTimes[i] != -1 {
			total += weights[i] * (weeksWorstTimes[i] + 1)
		} else if weeksTimes[i] != -1 {
			total += weights[i] * weeksTimes[i]
		}

		if weeksTimes[i] != -1 || weeksWorstTimes[i] != -1 {
			totalWeight += weights[i]
		}
	}

	average := int(float64(total) / float64(totalWeight))

	return average, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
