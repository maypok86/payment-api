package integration

import (
	"net/http"

	. "github.com/Eun/go-hit"
)

const (
	createOrderPath = basePath + "/order/create"
	payForOrderPath = basePath + "/order/pay"
	cancelOrderPath = basePath + "/order/cancel"
)

func (as *APISuite) TestCreateOrder() {
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
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusConflict),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Create order error. Order already exist",
		}),
	)

	Test(as.T(),
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 2,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Create order error. Account not found",
		}),
	)

	Test(as.T(),
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   -1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Create order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": -1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Create order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": -1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Create order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     -40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Create order error. Invalid request",
		}),
	)
}

func (as *APISuite) TestPayForOrder() {
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
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   2,
			"account_id": 1,
			"service_id": 1,
			"amount":     50,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".order").JQ(".order_id").Equal(2),
		Expect().Body().JSON().JQ(".order").JQ(".account_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".service_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".amount").Equal(50),
		Expect().Body().JSON().JQ(".order").JQ(".is_paid").Equal(false),
		Expect().Body().JSON().JQ(".order").JQ(".is_cancelled").Equal(false),
		Expect().Body().JSON().JQ(".balance").Equal(10),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Pay for order error. Order not found",
		}),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   3,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Pay for order error. Order not found",
		}),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   -1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Pay for order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": -1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Pay for order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": -1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Pay for order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(payForOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     -40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Pay for order error. Invalid request",
		}),
	)
}

func (as *APISuite) TestCancelOrder() {
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
		Post(createOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   2,
			"account_id": 1,
			"service_id": 1,
			"amount":     50,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".order").JQ(".order_id").Equal(2),
		Expect().Body().JSON().JQ(".order").JQ(".account_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".service_id").Equal(1),
		Expect().Body().JSON().JQ(".order").JQ(".amount").Equal(50),
		Expect().Body().JSON().JQ(".order").JQ(".is_paid").Equal(false),
		Expect().Body().JSON().JQ(".order").JQ(".is_cancelled").Equal(false),
		Expect().Body().JSON().JQ(".balance").Equal(10),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 50,
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Order not found",
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   3,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Order not found",
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 2,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Order not found",
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   -1,
			"account_id": 1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": -1,
			"service_id": 1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": -1,
			"amount":     40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(cancelOrderPath),
		Send().Body().JSON(map[string]interface{}{
			"order_id":   1,
			"account_id": 1,
			"service_id": 1,
			"amount":     -40,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Cancel order error. Invalid request",
		}),
	)
}
