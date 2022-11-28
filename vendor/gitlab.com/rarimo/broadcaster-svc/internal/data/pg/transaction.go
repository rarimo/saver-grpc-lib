package pg

import (
	"context"
	"github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/broadcaster-svc/internal/data"
)

func (q TransactionQ) Select(ctx context.Context) ([]data.Transaction, error) {
	stmt := squirrel.Select(colsTransaction).From("transactions")
	var result []data.Transaction
	err := q.db.SelectContext(ctx, &result, stmt)
	return result, errors.Wrap(err, "failed to exec stmt")
}
