package queries

import (
	"testing"
	"time"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/internal/helpers"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateStatusAndBonus(t *testing.T) {
	t.Parallel()

	t.Run("update status and bonus for PROCESSED order", func(t *testing.T) {
		orderID := "12345678903"
		status := entities.OrdersStatusProcessed
		bonusSum := 150.50

		stmt := NewUpdateStatusAndBonus(orderID, status, bonusSum)
		sql, args := stmt.Sql()

		expectedSQL := `UPDATE public.orders 
SET (status, bonus_sum, updated_at, processed_at) = ($1, $2, $3, $4) 
WHERE orders.order_id = $5::text;`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Len(t, args, 5)
		assert.Equal(t, model.OrdersStatus("PROCESSED"), args[0])
		assert.Equal(t, bonusSum, args[1])
		assert.IsType(t, time.Time{}, args[2]) // UpdatedAt
		assert.IsType(t, time.Time{}, args[3]) // ProcessedAt
		assert.Equal(t, orderID, args[4])
	})

	t.Run("update status and bonus for NEW order (without processed_at)", func(t *testing.T) {
		orderID := "12345678903"
		status := entities.OrdersStatusNew
		bonusSum := 0.0

		stmt := NewUpdateStatusAndBonus(orderID, status, bonusSum)
		sql, args := stmt.Sql()

		expectedSQL := `UPDATE public.orders 
SET (status, bonus_sum, updated_at) = ($1, $2, $3) 
WHERE orders.order_id = $4::text;`

		assert.Equal(t, helpers.NormalizeSQL(expectedSQL), helpers.NormalizeSQL(sql))
		assert.Len(t, args, 4)
		assert.Equal(t, model.OrdersStatus("NEW"), args[0])
		assert.Equal(t, bonusSum, args[1])
		assert.IsType(t, time.Time{}, args[2]) // UpdatedAt
		assert.Equal(t, orderID, args[3])
	})
}
