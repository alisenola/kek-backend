package database

import (
	"context"
	"fmt"
	"kek-backend/internal/alert/model"
	"kek-backend/internal/database"
	"kek-backend/pkg/logging"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IterateAlertCriteria struct {
	Account uint
	Offset  uint
	Limit   uint
}

//go:generate mockery --name AlertDB --filename alert_mock.go
type AlertDB interface {
	RunInTx(ctx context.Context, f func(ctx context.Context) error) error

	// SaveAlert saves a given alert.
	SaveAlert(ctx context.Context, alert *model.Alert) error

	// FindAlertBySlug returns a alert with given slug
	// database.ErrNotFound error is returned if not exist
	FindAlertBySlug(ctx context.Context, slug string) (*model.Alert, error)

	// FindAlerts returns alert list with given criteria and total count
	FindAlerts(ctx context.Context, criteria IterateAlertCriteria) ([]*model.Alert, int64, error)

	// FindAlertsWithoutContext returns alert list with given criteria and total count
	FindAlertsWithoutContext(criteria IterateAlertCriteria) ([]*model.Alert, int64, error)

	// DeleteAlertBySlug deletes a alert with given slug
	// and returns nil if success to delete, otherwise returns an error
	DeleteAlertBySlug(ctx context.Context, accountId uint, slug string) error
}

type alertDB struct {
	db *gorm.DB
}

func (a *alertDB) RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx := a.db.Begin()
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "start tx")
	}

	ctx = database.WithDB(ctx, tx)
	if err := f(ctx); err != nil {
		if err1 := tx.Rollback().Error; err1 != nil {
			return errors.Wrap(err, fmt.Sprintf("rollback tx: %v", err1.Error()))
		}
		return errors.Wrap(err, "invoke function")
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx: %v", err)
	}
	return nil
}

func (a *alertDB) SaveAlert(ctx context.Context, alert *model.Alert) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.SaveAlert", "alert", alert)

	if err := db.WithContext(ctx).Create(alert).Error; err != nil {
		logger.Errorw("alert.db.SaveAlert failed to save alert", "err", err)
		if database.IsKeyConflictErr(err) {
			return database.ErrKeyConflict
		}
		return err
	}
	return nil
}

func (a *alertDB) FindAlertBySlug(ctx context.Context, slug string) (*model.Alert, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.FindAlertBySlug", "slug", slug)

	var ret model.Alert
	// 1) load alert with account
	// SELECT alerts.*, accounts.*
	// FROM `alerts` LEFT JOIN `accounts` `Account` ON `alerts`.`account_id` = `Account`.`id`
	// WHERE slug = "title1" AND deleted_at_unix = 0 ORDER BY `alerts`.`id` LIMIT 1
	err := db.WithContext(ctx).Joins("Account").
		First(&ret, "slug = ? AND deleted_at_unix = 0", slug).Error

	if err != nil {
		logger.Errorw("failed to find alert", "err", err)
		if database.IsRecordNotFoundErr(err) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (a *alertDB) FindAlerts(ctx context.Context, criteria IterateAlertCriteria) ([]*model.Alert, int64, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.FindAlerts", "criteria", criteria)

	chain := db.WithContext(ctx).Table("alerts a").Where("deleted_at_unix = 0")
	if criteria.Account != 0 {
		chain = chain.Where("au.id = ?", criteria.Account).Joins("LEFT JOIN accounts au on au.id = a.account_id")
	}

	// get total count
	var totalCount int64
	err := chain.Distinct("a.id").Count(&totalCount).Error
	if err != nil {
		logger.Error("failed to get total count", "err", err)
	}

	// get alert ids
	rows, err := chain.Select("(a.id) id").
		Offset(int(criteria.Offset)).
		Limit(int(criteria.Limit)).
		Order("a.id DESC").
		Rows()
	if err != nil {
		logger.Error("failed to read alert ids", "err", err)
		return nil, 0, err
	}
	var ids []uint
	for rows.Next() {
		var id uint
		err := rows.Scan(&id)
		if err != nil {
			logger.Error("failed to scan id from id rows", "err", err)
			return nil, 0, err
		}
		ids = append(ids, id)
	}

	// get alerts with account by ids
	var ret []*model.Alert
	if len(ids) == 0 {
		return []*model.Alert{}, totalCount, nil
	}
	err = db.WithContext(ctx).Joins("Account").
		Where("alerts.id IN (?)", ids).
		Order("alerts.id DESC").
		Find(&ret).Error
	if err != nil {
		logger.Error("failed to find alert by ids", "err", err)
		return nil, 0, err
	}

	return ret, totalCount, nil
}

func (a *alertDB) FindAlertsWithoutContext(criteria IterateAlertCriteria) ([]*model.Alert, int64, error) {
	db := a.db

	chain := db.Table("alerts a").Where("deleted_at_unix = 0")
	if criteria.Account != 0 {
		chain = chain.Where("au.id = ?", criteria.Account).Joins("LEFT JOIN accounts au on au.id = a.account_id")
	}

	// get total count
	var totalCount int64
	err := chain.Distinct("a.id").Count(&totalCount).Error
	if err != nil {
	}

	// get alert ids
	rows, err := chain.Select("(a.id) id").
		Offset(int(criteria.Offset)).
		Limit(int(criteria.Limit)).
		Order("a.id DESC").
		Rows()
	if err != nil {
		return nil, 0, err
	}
	var ids []uint
	for rows.Next() {
		var id uint
		err := rows.Scan(&id)
		if err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}

	// get alerts with account by ids
	var ret []*model.Alert
	if len(ids) == 0 {
		return []*model.Alert{}, totalCount, nil
	}
	err = db.Joins("Account").
		Where("alerts.id IN (?)", ids).
		Order("alerts.id DESC").
		Find(&ret).Error
	if err != nil {
		return nil, 0, err
	}

	return ret, totalCount, nil
}

func (a *alertDB) DeleteAlertBySlug(ctx context.Context, accountId uint, slug string) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.DeleteAlertBySlug", "slug", slug)

	// delete alert
	chain := db.WithContext(ctx).Model(&model.Alert{}).
		Where("slug = ? AND deleted_at_unix = 0", slug).
		Where("account_id = ?", accountId).
		Update("deleted_at_unix", time.Now().Unix())
	if chain.Error != nil {
		logger.Errorw("failed to delete an alert", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		logger.Error("failed to delete an alert because not found")
		return database.ErrNotFound
	}
	return nil
}

// NewAlertDB creates a new alert db with given db
func NewAlertDB(db *gorm.DB) AlertDB {
	return &alertDB{
		db: db,
	}
}
