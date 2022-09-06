package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"../productpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Go client is running")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect %v", err)
	}

	defer cc.Close()

	c := productpb.NewProductServiceClient(cc)

	// productCreate(c)
	// productGet(c)
	// productsGet(c)

	//Creating Product

	product := &productpb.Product{
		Name:  "Laptop Lenovo",
		Price: 2500.50,
	}

	createdProduct, err := c.CreateProduct(context.Background(), &productpb.CreateProductRequest{
		Product: product,
	})

	if err != nil {
		log.Fatalf("Failed to create product %v", err)
	}

	fmt.Printf("Product created %v\n", createdProduct)

	//Getting Product

	productID := createdProduct.GetProduct().GetId()

	getProductReq := &productpb.GetProductRequest{ProductId: productID}

	getProductRes, err := c.GetProduct(context.Background(), getProductReq)

	if err != nil {
		log.Fatalf("Failed to getting product %v", err)
	}

	fmt.Printf("Product gitten: %v\n", getProductRes)

	//List Products

	stream, err := c.ListProduct(context.Background(), &productpb.ListProductRequest{})

	if err != nil {
		log.Fatalf("Error Calling List Product %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("error receiving product %v", err)
		}

		fmt.Println(res.GetProduct())
	}
}

// func productCreate(c productpb.ProductServiceClient) {
// 	product := &productpb.Product{
// 		Name:  "Laptop Lenovo",
// 		Price: 2500.50,
// 	}

// 	createdProduct, err := c.CreateProduct(context.Background(), &productpb.CreateProductRequest{
// 		Product: product,
// 	})

// 	if err != nil {
// 		log.Fatalf("Failed to create product %v", err)
// 	}

// 	fmt.Printf("Product created %v", createdProduct)
// }

// func productGet(c productpb.ProductServiceClient) {
// 	productID := createdProduct.GetProduct().GetId()

// 	getProductReq := &productpb.GetProductRequest{ProductId: productID}

// 	getProductRes, err := c.GetProduct(context.Background(), getProductReq)

// 	if err != nil {
// 		log.Fatalf("Failed to getting product %v", err)
// 	}

// 	fmt.Println("Product gitten: %v", getProductRes)
// }

// func productsGet(c productpb.ProductServiceClient) {
// 	stream, err := c.ListProduct(context.Background(), &productpb.ListProductRequest{})

// 	if err != nil {
// 		log.Fatalf("Error Calling List Product %v", err)
// 	}

// 	for {
// 		res, err := stream.Recv()
// 		if err == io.EOF {
// 			break
// 		}

// 		if err != nil {
// 			log.Fatalf("error receiving product %v", err)
// 		}

// 		fmt.Println(res.GetProduct())
// 	}
// }
