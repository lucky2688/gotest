package main

import (
	"errors"
	"fmt"
)

type BusinessInstance interface {
	ValidateLogin()
	ValidateParams()
	AntispamCheck()
	GetPrice()
	CreateOrder()
	UpdateUserStatus()
	NotifyDownstreamSystems()
}

type BusinessType string

const (
	TravelBusiness BusinessType = "travel"
	MarketBusiness BusinessType = "market"
)

// 示例：Travel 实现
type TravelOrder struct{}

func (TravelOrder) ValidateLogin()           { fmt.Println("Travel: ValidateLogin") }
func (TravelOrder) ValidateParams()          { fmt.Println("Travel: ValidateParams") }
func (TravelOrder) AntispamCheck()           { fmt.Println("Travel: AntispamCheck") }
func (TravelOrder) GetPrice()                { fmt.Println("Travel: GetPrice") }
func (TravelOrder) CreateOrder()             { fmt.Println("Travel: CreateOrder") }
func (TravelOrder) UpdateUserStatus()        { fmt.Println("Travel: UpdateUserStatus") }
func (TravelOrder) NotifyDownstreamSystems() { fmt.Println("Travel: NotifyDownstreamSystems") }

func NewTravelOrder() BusinessInstance {
	return &TravelOrder{}
}

// 示例：Market 实现
type MarketOrder struct{}

func (MarketOrder) ValidateLogin()           { fmt.Println("Market: ValidateLogin") }
func (MarketOrder) ValidateParams()          { fmt.Println("Market: ValidateParams") }
func (MarketOrder) AntispamCheck()           { fmt.Println("Market: AntispamCheck") }
func (MarketOrder) GetPrice()                { fmt.Println("Market: GetPrice") }
func (MarketOrder) CreateOrder()             { fmt.Println("Market: CreateOrder") }
func (MarketOrder) UpdateUserStatus()        { fmt.Println("Market: UpdateUserStatus") }
func (MarketOrder) NotifyDownstreamSystems() { fmt.Println("Market: NotifyDownstreamSystems") }

func NewMarketOrder() BusinessInstance {
	return &MarketOrder{}
}

func BusinessProcess(bi BusinessInstance) {
	bi.ValidateLogin()
	bi.ValidateParams()
	bi.AntispamCheck()
	bi.GetPrice()
	bi.CreateOrder()
	bi.UpdateUserStatus()
	bi.NotifyDownstreamSystems()
}

var businessInstanceFactory = map[string]func() BusinessInstance{
	string(TravelBusiness): NewTravelOrder,
	string(MarketBusiness): NewMarketOrder,
}

func entry(businessType BusinessType) error {
	//var bi BusinessInstance
	//switch businessType {
	//case TravelBusiness:
	//	bi = NewTravelOrder()
	//case MarketBusiness:
	//	bi = NewMarketOrder()
	//default:
	//	return errors.New("not supported business")
	//}
	factory, ok := businessInstanceFactory[string(businessType)]
	if !ok {
		return errors.New("not supported business")
	}
	bi := factory()
	BusinessProcess(bi)
	return nil
}

func main() {
	err := entry(TravelBusiness)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
