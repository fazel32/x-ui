package service

import (
	"fmt"
	"strings"
	"time"
	"x-ui/database"
	"encoding/json"
	"x-ui/database/model"
	"x-ui/util/common"
	"x-ui/xray"
	"x-ui/logger"

	"gorm.io/gorm"
)

type InboundService struct {
}

func (s *InboundService) GetInbounds(userId int) ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).Preload("ClientStats").Where("user_id = ?", userId).Find(&inbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return inbounds, nil
}

func (s *InboundService) GetAllInbounds() ([]*model.Inbound, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).Preload("ClientStats").Find(&inbounds).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return inbounds, nil
}

func (s *InboundService) checkPortExist(port int, ignoreId int) (bool, error) {
	db := database.GetDB()
	db = db.Model(model.Inbound{}).Where("port = ?", port)
	if ignoreId > 0 {
		db = db.Where("id != ?", ignoreId)
	}
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *InboundService) getClients(inbound *model.Inbound) ([]model.Client, error) {
	settings := map[string][]model.Client{}
	json.Unmarshal([]byte(inbound.Settings), &settings)
	if settings == nil {
		return nil, fmt.Errorf("Setting is null")
	}

	clients := settings["clients"]
	if clients == nil {
		return nil, nil
	}
	return clients, nil
}

func (s *InboundService) checkEmailsExist(emails map[string] bool, ignoreId int) (string, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound 
	db = db.Model(model.Inbound{}).Where("Protocol in ?", []model.Protocol{model.VMess, model.VLESS})
	if (ignoreId > 0) {
		db = db.Where("id != ?", ignoreId)
	}
	db = db.Find(&inbounds)
	if db.Error != nil {
		return "", db.Error
	}

	for _, inbound := range inbounds {
		clients, err := s.getClients(inbound)
		if err != nil {
			return "", err
		}

		for _, client := range clients {
			if emails[client.Email] {
				return client.Email, nil
			}
		}
	}
	return "", nil
}

func (s *InboundService) checkEmailExistForInbound(inbound *model.Inbound) (string, error) {
	clients, err := s.getClients(inbound)
	if err != nil {
		return "", err
	}
	emails := make(map[string] bool)
	for _, client := range clients {
		if (client.Email != "") {
			if emails[client.Email] {
				return client.Email, nil
			}
			emails[client.Email] = true;
		}
	}
	return s.checkEmailsExist(emails, inbound.Id)
}

func (s *InboundService) AddInbound(inbound *model.Inbound) (*model.Inbound,error) {
	exist, err := s.checkPortExist(inbound.Port, 0)
	if err != nil {
		return inbound, err
	}
	if exist {
		return inbound, common.NewError("Port already exists:", inbound.Port)
	}

	existEmail, err := s.checkEmailExistForInbound(inbound)
	if err != nil {
		return inbound, err
	}
	if existEmail != "" {
		return inbound, common.NewError("Duplicate email:", existEmail)
	}

	db := database.GetDB()

	err = db.Save(inbound).Error
	if err == nil {
		s.UpdateClientStat(inbound.Id,inbound.Settings)
	}
	return inbound, err
}

func (s *InboundService) AddInbounds(inbounds []*model.Inbound) error {
	for _, inbound := range inbounds {
		exist, err := s.checkPortExist(inbound.Port, 0)
		if err != nil {
			return err
		}
		if exist {
			return common.NewError("Port already exists:", inbound.Port)
		}
	}

	db := database.GetDB()
	tx := db.Begin()
	var err error
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	for _, inbound := range inbounds {
		err = tx.Save(inbound).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *InboundService) DelInbound(id int) error {
	db := database.GetDB()
	return db.Delete(model.Inbound{}, id).Error
}

func (s *InboundService) GetInbound(id int) (*model.Inbound, error) {
	db := database.GetDB()
	inbound := &model.Inbound{}
	err := db.Model(model.Inbound{}).First(inbound, id).Error
	if err != nil {
		return nil, err
	}
	return inbound, nil
}

func (s *InboundService) UpdateInbound(inbound *model.Inbound) (*model.Inbound, error) {
	exist, err := s.checkPortExist(inbound.Port, inbound.Id)
	if err != nil {
		return inbound, err
	}
	if exist {
		return inbound, common.NewError("Port already exists:", inbound.Port)
	}
	
	existEmail, err := s.checkEmailExistForInbound(inbound)
	if err != nil {
		return inbound, err
	}
	if existEmail != "" {
		return inbound, common.NewError("Duplicate email:", existEmail)
	}

	oldInbound, err := s.GetInbound(inbound.Id)
	if err != nil {
		return inbound, err
	}
	oldInbound.Up = inbound.Up
	oldInbound.Down = inbound.Down
	oldInbound.Total = inbound.Total
	oldInbound.Remark = inbound.Remark
	oldInbound.Enable = inbound.Enable
	oldInbound.ExpiryTime = inbound.ExpiryTime
	oldInbound.Listen = inbound.Listen
	oldInbound.Port = inbound.Port
	oldInbound.Protocol = inbound.Protocol
	oldInbound.Settings = inbound.Settings
	oldInbound.StreamSettings = inbound.StreamSettings
	oldInbound.Sniffing = inbound.Sniffing
	oldInbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)

	s.UpdateClientStat(inbound.Id,inbound.Settings)
	db := database.GetDB()
	return inbound, db.Save(oldInbound).Error
}

func (s *InboundService) AddTraffic(traffics []*xray.Traffic) (err error) {
	if len(traffics) == 0 {
		return nil
	}
	db := database.GetDB()
	db = db.Model(model.Inbound{})
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	for _, traffic := range traffics {
		if traffic.IsInbound {
			err = tx.Where("tag = ?", traffic.Tag).
				UpdateColumns(map[string]interface{}{
					"up":   gorm.Expr("up + ?", traffic.Up),
					"down": gorm.Expr("down + ?", traffic.Down)}).Error
			if err != nil {
				return
			}
		}
	}
	return
}

func (s *InboundService) AddClientTraffic(traffics []*xray.ClientTraffic) (err error) {
	if len(traffics) == 0 {
		return nil
	}
	db := database.GetDB()
	dbInbound := db.Model(model.Inbound{})

	db = db.Model(xray.ClientTraffic{})
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	txInbound := dbInbound.Begin()
	defer func() {
		if err != nil {
			txInbound.Rollback()
		} else {
			txInbound.Commit()
		}
	}()

	for _, traffic := range traffics {
		inbound := &model.Inbound{}
		client := &xray.ClientTraffic{}
		err := tx.Where("email = ?", traffic.Email).First(client).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Warning(err, traffic.Email)
			}
			continue
		}

		err = txInbound.Where("id=?", client.InboundId).First(inbound).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Warning(err, traffic.Email)

			}
			continue
		}
		// get settings clients
		settings := map[string][]model.Client{}
		json.Unmarshal([]byte(inbound.Settings), &settings)
		clients := settings["clients"]
		for _, client := range clients {
			if traffic.Email == client.Email {
				traffic.ExpiryTime = client.ExpiryTime
				traffic.Total = client.TotalGB
			}
		}
		if tx.Where("inbound_id = ? and email = ?", inbound.Id, traffic.Email).
			UpdateColumns(map[string]interface{}{
				"expiry_time": traffic.ExpiryTime,
				"total":       traffic.Total,
				"up":          gorm.Expr("up + ?", traffic.Up),
				"down":        gorm.Expr("down + ?", traffic.Down)}).RowsAffected == 0 {
			err = tx.Create(traffic).Error
		}

		if err != nil {
			logger.Warning("AddClientTraffic update data ", err)
			continue
		}

	}
	return
}

func (s *InboundService) DisableInvalidInbounds() (int64, error) {
	db := database.GetDB()
	now := time.Now().Unix() * 1000
	result := db.Model(model.Inbound{}).
		Where("((total > 0 and up + down >= total) or (expiry_time > 0 and expiry_time <= ?)) and enable = ?", now, true).
		Update("enable", false)
	err := result.Error
	count := result.RowsAffected
	return count, err
}
func (s *InboundService) DisableInvalidClients() (int64, error) {
	db := database.GetDB()
	now := time.Now().Unix() * 1000
	result := db.Model(xray.ClientTraffic{}).
		Where("((total > 0 and up + down >= total) or (expiry_time > 0 and expiry_time <= ?)) and enable = ?", now, true).
		Update("enable", false)
	err := result.Error
	count := result.RowsAffected
	return count, err
}

func (s *InboundService) RemoveOrphanedTraffics() {
	db := database.GetDB()
	db.Exec(`
		DELETE FROM client_traffics
		WHERE email NOT IN (
			SELECT JSON_EXTRACT(client.value, '$.email')
			FROM inbounds,
				JSON_EACH(JSON_EXTRACT(inbounds.settings, '$.clients')) AS client
		)
	`)
}

func (s *InboundService) UpdateClientStat(inboundId int, inboundSettings string) (error) {
	db := database.GetDB()

	// get settings clients
	settings := map[string][]model.Client{}
	json.Unmarshal([]byte(inboundSettings), &settings)
	clients := settings["clients"]
	for _, client := range clients {
		result := db.Model(xray.ClientTraffic{}).
		Where("inbound_id = ? and email = ?", inboundId, client.Email).
		Updates(map[string]interface{}{"total": client.TotalGB, "expiry_time": client.ExpiryTime})
		if result.RowsAffected == 0 {
			clientTraffic := xray.ClientTraffic{}
			clientTraffic.InboundId = inboundId
			clientTraffic.Email = client.Email
			clientTraffic.Total = client.TotalGB
			clientTraffic.ExpiryTime = client.ExpiryTime
			clientTraffic.Enable = true
			clientTraffic.Up = 0
			clientTraffic.Down = 0
			db.Create(&clientTraffic)
		}
		err := result.Error
		if err != nil {
			return err
		}
	
	}
	return nil
}
func (s *InboundService) DelClientStat(tx *gorm.DB, email string) error {
	return tx.Where("email = ?", email).Delete(xray.ClientTraffic{}).Error
}

func (s *InboundService) GetInboundClientIps(clientEmail string) (string, error) {
	db := database.GetDB()
	InboundClientIps := &model.InboundClientIps{}
	err := db.Model(model.InboundClientIps{}).Where("client_email = ?", clientEmail).First(InboundClientIps).Error
	if err != nil {
		return "", err
	}
	return InboundClientIps.Ips, nil
}
func (s *InboundService) ClearClientIps(clientEmail string) (error) {
	db := database.GetDB()

	result := db.Model(model.InboundClientIps{}).
		Where("client_email = ?", clientEmail).
		Update("ips", "")
	err := result.Error


	if err != nil {
		return err
	}
	return nil
}
func (s *InboundService) ResetClientTraffic(id int, clientEmail string) error {
	db := database.GetDB()

	result := db.Model(xray.ClientTraffic{}).
		Where("inbound_id = ? and email = ?", id, clientEmail).
		Updates(map[string]interface{}{"enable": true, "up": 0, "down": 0})

	err := result.Error

	if err != nil {
		return err
	}
	return nil
}

func (s *InboundService) ResetAllClientTraffics(id int) error {
	db := database.GetDB()

	whereText := "inbound_id "
	if id == -1 {
		whereText += " > ?"
	} else {
		whereText += " = ?"
	}

	result := db.Model(xray.ClientTraffic{}).
		Where(whereText, id).
		Updates(map[string]interface{}{"enable": true, "up": 0, "down": 0})

	err := result.Error

	if err != nil {
		return err
	}
	return nil
}

func (s *InboundService) ResetAllTraffics() error {
	db := database.GetDB()

	result := db.Model(model.Inbound{}).
		Where("user_id > ?", 0).
		Updates(map[string]interface{}{"up": 0, "down": 0})

	err := result.Error

	if err != nil {
		return err
	}
	return nil
}

func (s *InboundService) DelDeactiveClients(id int) (err error) {
	db := database.GetDB()
	tx := db.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	whereText := "inbound_id "
	if id < 0 {
		whereText += "> ?"
	} else {
		whereText += "= ?"
	}

	deactiveClients := []xray.ClientTraffic{}
	err = db.Model(xray.ClientTraffic{}).Where(whereText+" and enable = ?", id, false).Select("inbound_id, GROUP_CONCAT(email) as email").Group("inbound_id").Find(&deactiveClients).Error
	if err != nil {
		return err
	}

	for _, deactiveClient := range deactiveClients {
		emails := strings.Split(deactiveClient.Email, ",")
		oldInbound, err := s.GetInbound(deactiveClient.InboundId)
		if err != nil {
			return err
		}
		var oldSettings map[string]interface{}
		err = json.Unmarshal([]byte(oldInbound.Settings), &oldSettings)
		if err != nil {
			return err
		}

		oldClients := oldSettings["clients"].([]interface{})
		var newClients []interface{}
		for _, client := range oldClients {
			deplete := false
			c := client.(map[string]interface{})
			for _, email := range emails {
				if email == c["email"].(string) {
					deplete = true
					break
				}
			}
			if !deplete {
				newClients = append(newClients, client)
			}
		}
		if len(newClients) > 0 {
			oldSettings["clients"] = newClients

			newSettings, err := json.MarshalIndent(oldSettings, "", "  ")
			if err != nil {
				return err
			}

			oldInbound.Settings = string(newSettings)
			err = tx.Save(oldInbound).Error
			if err != nil {
				return err
			}
		} else {
			// Delete inbound if no client remains
			s.DelInbound(deactiveClient.InboundId)
		}
	}

	err = tx.Where(whereText+" and enable = ?", id, false).Delete(xray.ClientTraffic{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *InboundService) SwitchClientStatus(clientEmail string) (error) {
  db := database.GetDB()
  
  client := &xray.ClientTraffic{}
  err :=db.Model(xray.ClientTraffic{}).
		Where("email = ?", clientEmail).
    First(client).
		Error

  if err != nil {
		return err
	}

  switchStatus := !client.Enable
	result := db.Model(xray.ClientTraffic{}).
		Where("email = ?", clientEmail).
		Update("enable", switchStatus  ) 

	err = result.Error

	if err != nil {
		return err
	}
	return nil

}
