package db

import "gorm.io/gorm"

func CreatePlanDraft(r *PlanDraftRecord) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(r).Error
}

func GetPlanDraft(id string) (*PlanDraftRecord, error) {
	if DB == nil {
		return nil, gorm.ErrRecordNotFound
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var r PlanDraftRecord
	err := DB.WithContext(ctx).Where("id = ?", id).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func ListPlanDraftsByUser(userID uint) ([]PlanDraftRecord, error) {
	if DB == nil {
		return nil, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var records []PlanDraftRecord
	err := DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func UpdatePlanDraft(id string, updates map[string]any) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Model(&PlanDraftRecord{}).Where("id = ?", id).Updates(updates).Error
}
