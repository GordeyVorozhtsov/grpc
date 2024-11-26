package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	pb "proto" // Измените путь на правильный

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedCurrencyExchangerServer
	db *sql.DB
}

func (s *server) GetExchangeRate(ctx context.Context, req *pb.ExchangeRateRequest) (*pb.ExchangeRateResponse, error) {
	var rate float64

	// Запрашиваем курс из базы данных
	err := s.db.QueryRow("SELECT rate FROM currencies WHERE code = $1", req.CurrencyPair).Scan(&rate)
	if err != nil {
		return nil, err
	}

	return &pb.ExchangeRateResponse{Rate: rate}, nil
}

func (s *server) ConvertCurrency(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	var rate float64
	currencyPair := req.FromCurrency + "/" + req.ToCurrency

	// Получаем курс валюты
	err := s.db.QueryRow("SELECT rate FROM currencies WHERE code = $1", currencyPair).Scan(&rate)
	if err != nil {
		return nil, err
	}

	convertedAmount := req.Amount * rate
	return &pb.ConvertResponse{ConvertedAmount: convertedAmount}, nil
}

func main() {
	db, err := sql.Open("postgres", "user=fucku dbname=currencies sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCurrencyExchangerServer(grpcServer, &server{db: db})

	log.Println("Starting gRPC server on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
