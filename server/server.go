package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"../productpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type product struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name"`
	Price float64            `bson:"price"`
}

type server struct{}

func dbToProductPb(data *product) *productpb.Product {
	return &productpb.Product{
		Id:    data.ID.Hex(),
		Name:  data.Name,
		Price: data.Price,
	}
}

func (*server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	fmt.Println("Creating Products")
	prod := req.GetProduct()
	data := product{
		Name:  prod.GetName(),
		Price: prod.GetPrice(),
	}

	res, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Error: %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)

	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannon convert OID Error: %v", err),
		)
	}

	return &productpb.CreateProductResponse{
		Product: &productpb.Product{
			Id:    oid.Hex(),
			Name:  prod.GetName(),
			Price: prod.GetPrice(),
		},
	}, nil
}

func (*server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	productId := req.GetProductId()
	oid, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprint("Cannot parse ID"),
		)
	}

	data := &product{}

	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)

	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find the product: %v", err),
		)
	}

	return &productpb.GetProductResponse{
		Product: dbToProductPb(data),
	}, nil
}

func (*server) ListProduct(req *productpb.ListProductRequest, stream productpb.ProductService_ListProductServer) error {
	fmt.Println("List Products")

	cur, err := collection.Find(context.Background(), primitive.D{{}})

	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &product{}

		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error decoding data product: %v", err),
			)
		}

		stream.Send(&productpb.ListProductResponse{
			Product: dbToProductPb(data),
		})
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error decoding data product: %v", err),
		)
	}

	return nil
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error creating the client DB %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connecting the client DB %v", err)
	}

	collection = client.Database("productsdb").Collection("products")

	fmt.Print("Running go Server\n")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to Listen %v", err)
	}

	s := grpc.NewServer()

	productpb.RegisterProductServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting Server")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping the server")

	s.Stop()
	fmt.Println("Closing the listener")
	client.Disconnect(context.TODO())
	fmt.Println("Closing mongo client connection")
	lis.Close()
	fmt.Println("Goodbye... ")
}
