package dao

import (
	"github.com/ninjadotorg/handshake-ethereum/models"
	"log"
	"github.com/jinzhu/gorm"
	"time"
	"strings"
)

type EthereumTransactionsDao struct {
}

func (contractLogsDao EthereumTransactionsDao) GetById(id int64) (models.EthereumTransactions) {
	dto := models.EthereumTransactions{}
	err := models.Database().Where("id = ?", id).First(&dto).Error
	if err != nil {
		log.Print(err)
	}
	return dto
}

func (contractLogsDao EthereumTransactionsDao) GetByFilter(contractAddress string, event string) (models.EthereumTransactions) {
	contractAddress = strings.ToLower(contractAddress)
	dto := models.EthereumTransactions{}
	err := models.Database().Where("contract_address = ? AND event = ?", contractAddress, event).Order("block_number desc").First(&dto).Error
	if err != nil {
		log.Print(err)
	}
	return dto
}

func (contractLogsDao EthereumTransactionsDao) Create(dto models.EthereumTransactions, tx *gorm.DB) (models.EthereumTransactions, error) {
	if tx == nil {
		tx = models.Database()
	}
	dto.FromAddress = strings.ToLower(dto.FromAddress)
	dto.ToAddress = strings.ToLower(dto.ToAddress)
	dto.Hash = strings.ToLower(dto.Hash)
	dto.DateCreated = time.Now()
	dto.DateModified = dto.DateCreated
	err := tx.Create(&dto).Error
	if err != nil {
		log.Println(err)
		return dto, err
	}
	return dto, nil
}

func (contractLogsDao EthereumTransactionsDao) Update(dto models.EthereumTransactions, tx *gorm.DB) (models.EthereumTransactions, error) {
	if tx == nil {
		tx = models.Database()
	}
	dto.FromAddress = strings.ToLower(dto.FromAddress)
	dto.ToAddress = strings.ToLower(dto.ToAddress)
	dto.Hash = strings.ToLower(dto.Hash)
	dto.DateModified = time.Now()
	err := tx.Save(&dto).Error
	if err != nil {
		log.Println(err)
		return dto, err
	}
	return dto, nil
}

func (contractLogsDao EthereumTransactionsDao) Delete(dto models.EthereumTransactions, tx *gorm.DB) (models.EthereumTransactions, error) {
	if tx == nil {
		tx = models.Database()
	}
	err := tx.Delete(&dto).Error
	if err != nil {
		log.Println(err)
		return dto, err
	}
	return dto, nil
}
