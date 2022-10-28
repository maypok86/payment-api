package integration

import (
	"net/http"

	. "github.com/Eun/go-hit"
)

const (
	getTransactionsByAccountIDPath = basePath + "/transaction/"
)

func (as *APISuite) TestGetTransactionsByAccountID() {
	Test(as.T(),
		Post(addBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"account_id": 1,
			"amount":     100,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 100,
		}),
	)

	Test(as.T(),
		Post(addBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"account_id": 2,
			"amount":     50,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 50,
		}),
	)

	Test(as.T(),
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".order").JQ(".order_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".account_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".service_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".amount").Equal(40),
		Expect().Body().JSON().JQ(".order").JQ(".is_paid").Equal(false),
		Expect().Body().JSON().JQ(".order").JQ(".is_cancelled").Equal(false),
		Expect().Body().JSON().JQ(".balance").Equal(60),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   2,
			"receiver_id": 1,
			"amount":      30,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"sender_balance":   20,
			"receiver_balance": 90,
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".type").Equal("enrollment"),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".sender_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".amount").Equal(100),
		Expect().Body().
			JSON().
			JQ(".transactions.[0]").
			JQ(".description").
			Equal("Add 100 kopecks to account with id = 1"),

		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".type").Equal("reservation"),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".sender_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".amount").Equal(40),
		Expect().Body().
			JSON().
			JQ(".transactions.[1]").
			JQ(".description").
			Equal("Reserve 40 kopecks for order with id = 1"),

		Expect().Body().JSON().JQ(".transactions.[2]").JQ(".type").Equal("transfer"),
		Expect().Body().JSON().JQ(".transactions.[2]").JQ(".sender_id").Equal(2),
		Expect().Body().JSON().JQ(".transactions.[2]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[2]").JQ(".amount").Equal(30),
		Expect().Body().
			JSON().
			JQ(".transactions.[2]").
			JQ(".description").
			Equal("Transfer 30 kopecks from account with id = 2 to account with id = 1"),

		Expect().Body().JSON().JQ(".range").Equal(map[string]interface{}{
			"limit":  10,
			"offset": 0,
			"count":  3,
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?limit=2"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".type").Equal("enrollment"),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".sender_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".amount").Equal(100),
		Expect().Body().
			JSON().
			JQ(".transactions.[0]").
			JQ(".description").
			Equal("Add 100 kopecks to account with id = 1"),

		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".type").Equal("reservation"),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".sender_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".amount").Equal(40),
		Expect().Body().
			JSON().
			JQ(".transactions.[1]").
			JQ(".description").
			Equal("Reserve 40 kopecks for order with id = 1"),

		Expect().Body().JSON().JQ(".range").Equal(map[string]interface{}{
			"limit":  2,
			"offset": 0,
			"count":  3,
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?limit=2&offset=1"),
		Expect().Status().Equal(http.StatusOK),

		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".type").Equal("reservation"),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".sender_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".amount").Equal(40),
		Expect().Body().
			JSON().
			JQ(".transactions.[0]").
			JQ(".description").
			Equal("Reserve 40 kopecks for order with id = 1"),

		Expect().Body().JSON().JQ(".range").Equal(map[string]interface{}{
			"limit":  2,
			"offset": 1,
			"count":  3,
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?limit=2&offset=0&sort=sum&direction=asc"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".type").Equal("transfer"),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".sender_id").Equal(2),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".amount").Equal(30),
		Expect().Body().
			JSON().
			JQ(".transactions.[0]").
			JQ(".description").
			Equal("Transfer 30 kopecks from account with id = 2 to account with id = 1"),

		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".type").Equal("reservation"),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".sender_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".amount").Equal(40),
		Expect().Body().
			JSON().
			JQ(".transactions.[1]").
			JQ(".description").
			Equal("Reserve 40 kopecks for order with id = 1"),

		Expect().Body().JSON().JQ(".range").Equal(map[string]interface{}{
			"limit":  2,
			"offset": 0,
			"count":  3,
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"2"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".type").Equal("enrollment"),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".sender_id").Equal(2),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".receiver_id").Equal(2),
		Expect().Body().JSON().JQ(".transactions.[0]").JQ(".amount").Equal(50),
		Expect().Body().
			JSON().
			JQ(".transactions.[0]").
			JQ(".description").
			Equal("Add 50 kopecks to account with id = 2"),

		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".type").Equal("transfer"),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".sender_id").Equal(2),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".receiver_id").Equal(1),
		Expect().Body().JSON().JQ(".transactions.[1]").JQ(".amount").Equal(30),
		Expect().Body().
			JSON().
			JQ(".transactions.[1]").
			JQ(".description").
			Equal("Transfer 30 kopecks from account with id = 2 to account with id = 1"),

		Expect().Body().JSON().JQ(".range").Equal(map[string]interface{}{
			"limit":  10,
			"offset": 0,
			"count":  2,
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"-1"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Transactions not found. id is not valid",
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"invalid"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Transactions not found. id is not valid",
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?limit=-1"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Transactions not found. Pagination params is not valid",
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?offset=-1"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Transactions not found. Pagination params is not valid",
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?sort=sam"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Transactions not found. Sort param is not valid",
		}),
	)

	Test(
		as.T(),
		Get(getTransactionsByAccountIDPath+"1?sort=date&direction=ask"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Transactions not found. Direction param is not valid",
		}),
	)
}
