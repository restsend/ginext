package ginext

import (
	"errors"
	"time"
)

func (um *UserManager) MakeToken(user *GinExtUser) (obj GinToken, err error) {
	tx := um.db.Model(&GinToken{})
	token := GenUniqueKey(tx, "token", defaultTokenLength)
	obj = GinToken{
		CreatedAt: time.Now(),
		OwnerID:   user.ID,
		Token:     token,
		ExpiredAt: time.Now().Add(um.TokenExpired),
	}
	result := um.db.Create(&obj)
	if result.Error != nil {
		return obj, result.Error
	}
	return obj, nil
}

func (um *UserManager) GetUserByToken(token string) (obj *GinToken, err error) {
	tx := um.db.Where("token", token).Preload("Owner")
	result := tx.Take(&obj)
	if result.Error != nil {
		return nil, result.Error
	}
	if time.Since(obj.ExpiredAt) > 0 {
		tx.Delete(&GinToken{})
		return nil, errors.New("token expired")
	}
	return obj, nil
}

func (um *UserManager) TouchToken(tokenId uint) (err error) {
	result := um.db.Model(&GinToken{}).Where("id", tokenId).UpdateColumn("ExpiredAt", time.Now().Add(um.TokenExpired))
	return result.Error
}

func (um *UserManager) DeleteToken(token string) (err error) {
	result := um.db.Where("token", token).Delete(GinToken{})
	return result.Error
}
