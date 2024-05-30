package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", client)
	for _, parcel := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
	}
	fmt.Println()

	return nil
}

func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", number, nextStatus)

	return s.store.SetStatus(number, nextStatus)
}

func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Fatalf("connection to db failed: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db) // создайте объект ParcelStore функцией NewParcelStore
	service := NewParcelService(store)

	// регистрация посылки
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Printf("register service failed: %v\n", err)
		return
	}

	// изменение адреса
	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	if err = service.ChangeAddress(p.Number, newAddress); err != nil {
		fmt.Printf("changing adress failed: %v\n", err)
		return
	}

	// изменение статуса
	if err = service.NextStatus(p.Number); err != nil {
		fmt.Printf("next status failed: %v\n", err)
		return
	}

	// вывод посылок клиента
	if err = service.PrintClientParcels(client); err != nil {
		fmt.Printf("client print failed: %v\n", err)
		return
	}

	// попытка удаления отправленной посылки
	if err = service.Delete(p.Number); err != nil {
		fmt.Printf("failed to delete parcel: %v\n", err)
		return
	}

	// вывод посылок клиента
	// предыдущая посылка не должна удалиться, т.к. её статус НЕ «зарегистрирована»
	if err = service.PrintClientParcels(client); err != nil {
		fmt.Printf("failed to print client parcels: %v\n", err)
		return
	}

	// регистрация новой посылки
	if p, err = service.Register(client, address); err != nil {
		fmt.Printf("register new parcel failed: %v\n", err)
		return
	}

	// удаление новой посылки
	if err = service.Delete(p.Number); err != nil {
		fmt.Printf("failed to delete new parcel: %v\n", err)
		return
	}

	// вывод посылок клиента
	// здесь не должно быть последней посылки, т.к. она должна была успешно удалиться
	if err = service.PrintClientParcels(client); err != nil {
		fmt.Printf("failed to print client parcels after delete: %v\n", err)
		return
	}
}
