// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"bytes"
	"time"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	pb "github.com/mukundha/hipster-webapp/frontend/genproto"	
	"github.com/mukundha/hipster-webapp/frontend/genproto"
	"github.com/pkg/errors"
)

const (
	avoidNoopCurrencyConversionRPC = false
)

func (fe *frontendServer) getCurrencies(ctx context.Context) ([]string, error) {
	fmt.Println("getCurrencies")
	response, err := http.Get(fe.apiBaseUrl + "/currencies")
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)    
    var currs hipstershop.GetSupportedCurrenciesResponse
	if err := json.Unmarshal(data,&currs); err != nil {
			fmt.Println(err)
	}	
	var out []string	
	for _, c := range currs.CurrencyCodes {
		if _, ok := whitelistedCurrencies[c]; ok {
			out = append(out, c)
		}
	}	
	return out, nil
}

func (fe *frontendServer) getProducts(ctx context.Context) ([]*pb.Product, error) {
	fmt.Println("getProducts")
	response, err := http.Get(fe.apiBaseUrl + "/products")
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)	
	var list hipstershop.ListProductsResponse
	if err := json.Unmarshal(data,&list); err != nil {
			fmt.Println(err)
	}		
	return list.GetProducts(), err
}

func (fe *frontendServer) getProduct(ctx context.Context, id string) (*pb.Product, error) {
	fmt.Println("getProduct")
	response, err := http.Get(fe.apiBaseUrl + "/products/" + id )
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
	data, _ := ioutil.ReadAll(response.Body)		
	var product *hipstershop.Product
	if err := json.Unmarshal(data,&product); err != nil {
			fmt.Println(err)
	}				
	return product, err
}

func (fe *frontendServer) getCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	fmt.Println("getCart")
	response, err := http.Get(fe.apiBaseUrl + "/carts/" + userID)
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(data))
	var cart hipstershop.Cart
	if err := json.Unmarshal(data,&cart); err != nil {
			fmt.Println(err)
	}		
	return cart.GetItems(), err
}

func (fe *frontendServer) emptyCart(ctx context.Context, userID string) error {
	req, err := http.NewRequest("DELETE", fe.apiBaseUrl + "/carts/" + userID , nil)	
	response, err := http.DefaultClient.Do(req)	
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(data)
	return err
}

func (fe *frontendServer) insertCart(ctx context.Context, userID, productID string, quantity int32) error {
	fmt.Println("Insert Cart")
	req := hipstershop.AddItemRequest{
		UserId: userID,
		Item: &hipstershop.CartItem{
			ProductId: productID,
			Quantity:  quantity},
	 }
	var s,err = json.Marshal(req)
	if err != nil {
		fmt.Println("error:", err)
		return err
	}
	fmt.Println(string(s))
	response,err := http.Post(fe.apiBaseUrl + "/carts","application/json",bytes.NewBuffer(s))
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
	}    
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(data)) 
	return err
}

func (fe *frontendServer) convertCurrency(ctx context.Context, money *pb.Money, currency string) (*pb.Money, error) {
	fmt.Println("Converting Currency " +  money.GetCurrencyCode() + ", " + currency)
	if avoidNoopCurrencyConversionRPC && money.GetCurrencyCode() == currency {
		return money, nil
	}
	if (currency == "" || money.GetCurrencyCode() == "") {
		return money, nil
	}	
	query := fmt.Sprintf("%s%d", "?from.units=" , money.GetUnits())
	response, err := http.Get(fe.apiBaseUrl + "/currencies/convert/" + money.GetCurrencyCode() + "/" + currency + query)
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)	
	var conv hipstershop.Money
	if err := json.Unmarshal(data,&conv); err != nil {
			fmt.Println(err)
	}			
	return &conv, err
}

func (fe *frontendServer) getShippingQuote(ctx context.Context, items []*pb.CartItem, currency string) (*pb.Money, error) {
	fmt.Printf("Get Shipping Quote: %s ", currency)
	req := hipstershop.GetQuoteRequest{
		Address: nil,
		Items:   items}

	var s,err = json.Marshal(req)
	if err != nil {
		fmt.Println("error:", err)
		return nil,err
	}
	fmt.Printf("Shipping Quote Request: %s" , s )
	response, err := http.Post(fe.apiBaseUrl + "/shipping/quote","application/json",bytes.NewBuffer(s))
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)	
	fmt.Printf("Shipping Quote Response: %s", data)	
	var quote hipstershop.GetQuoteResponse
	if err := json.Unmarshal(data,&quote); err != nil {
		fmt.Println(err)
		return nil,err
	}		
	localized, err := fe.convertCurrency(ctx, quote.GetCostUsd(), currency)
	return localized, errors.Wrap(err, "failed to convert currency for shipping cost")
}

func (fe *frontendServer) getRecommendations(ctx context.Context, userID string, productIDs []string) ([]*pb.Product, error) {	
	
	//TODO: Include Products for recommendations
	response, err := http.Get(fe.apiBaseUrl + "/recommendations/" + userID)
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)	
	
	var resp hipstershop.ListRecommendationsResponse
	if err := json.Unmarshal(data,&resp); err != nil {
		fmt.Println(err)
		return nil,err
	}		
	
	out := make([]*pb.Product, len(resp.GetProductIds()))
	for i, v := range resp.GetProductIds() {
		p, err := fe.getProduct(ctx, v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get recommended product info (#%s)", v)
		}
		out[i] = p
	}
	if len(out) > 4 {
		out = out[:4] // take only first four to fit the UI
	}
	return out, err
}

func (fe *frontendServer) getAd(ctx context.Context, ctxKeys []string) ([]*pb.Ad, error) {	
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	//TODO: Use ContextKeys
	response, err := http.Get(fe.apiBaseUrl + "/ads")
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
    data, _ := ioutil.ReadAll(response.Body)	
	var ads hipstershop.AdResponse
	if err := json.Unmarshal(data,&ads); err != nil {
			fmt.Println(err)
	}		
	defer cancel()	
	return ads.GetAds(), errors.Wrap(err, "failed to get ads")
}

func (fe *frontendServer) placeOrder(ctx context.Context, orderRequest *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	var s,err = json.Marshal(orderRequest)
	if err != nil {
		fmt.Println("error:", err)
		return nil,err
	}
	fmt.Printf("Parsed Order JSON: %s", s)
	response, err := http.Post(fe.apiBaseUrl + "/orders/checkout","application/json",bytes.NewBuffer(s))
    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } 
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("Parse Order response: %s", data)	
	var orderResponse *hipstershop.PlaceOrderResponse
	if err := json.Unmarshal(data,&orderResponse); err != nil {
		fmt.Println(err)
	}	
	return orderResponse,err
}