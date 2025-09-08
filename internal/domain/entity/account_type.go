package entity

type AccountType int

const (
	AccountTypeBank AccountType = iota + 1
	AccountTypeCreditCard
)
