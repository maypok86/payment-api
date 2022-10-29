package integration

import (
	"net/http"

	. "github.com/Eun/go-hit"
)

const (
	addBalancePath      = basePath + "/balance/add"
	getBalancePath      = basePath + "/balance/"
	transferBalancePath = basePath + "/balance/transfer"
)

func (as *APISuite) TestAddBalance() {
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
			"account_id": 1,
			"amount":     100,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 200,
		}),
	)

	Test(as.T(),
		Post(addBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"account_id": 0,
			"amount":     100,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Amount not added. request is not valid",
		}),
	)

	Test(as.T(),
		Post(addBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"account_id": 1,
			"amount":     -1,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Amount not added. request is not valid",
		}),
	)
}

func (as *APISuite) TestGetBalance() {
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
		Get(getBalancePath+"1"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 100,
		}),
	)

	Test(as.T(),
		Get(getBalancePath+"2"),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Get balance by error. Account not found",
		}),
	)

	Test(as.T(),
		Get(getBalancePath+"-1"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Balance not found. id is not valid",
		}),
	)

	Test(as.T(),
		Get(getBalancePath+"invalid"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Balance not found. id is not valid",
		}),
	)
}

func (as *APISuite) TestTransferBalance() {
	Test(as.T(),
		Post(addBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"account_id": 1,
			"amount":     200,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 200,
		}),
	)

	Test(as.T(),
		Post(addBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"account_id": 2,
			"amount":     100,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"balance": 100,
		}),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   1,
			"receiver_id": 2,
			"amount":      100,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"sender_balance":   100,
			"receiver_balance": 200,
		}),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   3,
			"receiver_id": 2,
			"amount":      100,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Sender or receiver not found",
		}),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   -1,
			"receiver_id": 2,
			"amount":      100,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Amount not transferred. request is not valid",
		}),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   1,
			"receiver_id": -2,
			"amount":      100,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Amount not transferred. request is not valid",
		}),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   1,
			"receiver_id": 2,
			"amount":      0,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Amount not transferred. request is not valid",
		}),
	)

	Test(as.T(),
		Post(transferBalancePath),
		Send().Body().JSON(map[string]interface{}{
			"sender_id":   1,
			"receiver_id": 2,
			"amount":      -1,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Amount not transferred. request is not valid",
		}),
	)
}
