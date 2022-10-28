package integration

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	getReportLinkPath = basePath + "/report/link"
)

func (as *APISuite) generateRandomReport(accountNumber, serviceNumber, orderNumber, initBalance int) []int {
	rand.Seed(time.Now().UnixNano())

	accountBalances := make([]int, accountNumber+1)
	serviceAmounts := make([]int, serviceNumber+1)

	for i := 1; i <= accountNumber; i++ {
		Test(as.T(),
			Post(addBalancePath),
			Send().Body().JSON(map[string]interface{}{
				"account_id": i,
				"amount":     initBalance,
			}),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().Equal(map[string]interface{}{
				"balance": initBalance,
			}),
		)
		accountBalances[i] = initBalance
	}

	maxAmount := initBalance / orderNumber
	for orderID := 1; orderID <= orderNumber; orderID++ {
		accountID := rand.Intn(accountNumber) + 1
		serviceID := rand.Intn(serviceNumber) + 1
		amount := rand.Intn(maxAmount) + 1
		accountBalances[accountID] -= amount
		serviceAmounts[serviceID] += amount

		Test(as.T(),
			Post(createOrderPath),
			Send().Body().JSON(map[string]interface{}{
				"order_id":   orderID,
				"account_id": accountID,
				"service_id": serviceID,
				"amount":     amount,
			}),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".order").JQ(".order_id").Equal(orderID),
			Expect().Body().JSON().JQ(".order").JQ(".account_id").Equal(accountID),
			Expect().Body().JSON().JQ(".order").JQ(".service_id").Equal(serviceID),
			Expect().Body().JSON().JQ(".order").JQ(".amount").Equal(amount),
			Expect().Body().JSON().JQ(".order").JQ(".is_paid").Equal(false),
			Expect().Body().JSON().JQ(".order").JQ(".is_cancelled").Equal(false),
			Expect().Body().JSON().JQ(".balance").Equal(accountBalances[accountID]),
		)

		Test(as.T(),
			Post(payForOrderPath),
			Send().Body().JSON(map[string]interface{}{
				"order_id":   orderID,
				"account_id": accountID,
				"service_id": serviceID,
				"amount":     amount,
			}),
			Expect().Status().Equal(http.StatusOK),
		)
	}

	return serviceAmounts
}

func (as *APISuite) serviceAmountsToCSV(serviceAmounts []int) []byte {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	as.Require().NoError(writer.Write([]string{"service_id", "amount"}))
	for serviceID := 1; serviceID < len(serviceAmounts); serviceID++ {
		amount := serviceAmounts[serviceID]
		if amount > 0 {
			as.Require().NoError(writer.Write([]string{fmt.Sprintf("%d", serviceID), fmt.Sprintf("%d", amount)}))
		}
	}
	writer.Flush()

	as.Require().NoError(writer.Error())

	return buffer.Bytes()
}

func getNextMonth(month time.Month) time.Month {
	if month == time.December {
		return time.January
	}
	return month + 1
}

func (as *APISuite) TestGetReportLink() {
	year, month, _ := time.Now().Date()

	as.generateRandomReport(10, 5, 20, 20000)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": month,
			"year":  year,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"link": basePath + fmt.Sprintf("/report/?key=%d-%d", year, month),
		}),
	)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": getNextMonth(month),
			"year":  year,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Get report link error. Report is not available",
		}),
	)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": month,
			"year":  year + 1,
		}),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Get report link error. Report is not available",
		}),
	)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": 0,
			"year":  2022,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Get report link error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": 13,
			"year":  2022,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Get report link error. Invalid request",
		}),
	)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": 1,
			"year":  2021,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Get report link error. Invalid request",
		}),
	)
}

func (as *APISuite) TestDownloadReport() {
	year, month, _ := time.Now().Date()
	key := fmt.Sprintf("%d-%d", year, month)
	link := basePath + fmt.Sprintf("/report/?key=%s", key)
	reportFilename := fmt.Sprintf("report_%s.csv", key)
	Test(as.T(),
		Get(link),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Download report error. Report not found",
		}),
	)

	Test(as.T(),
		Get(basePath+"/report/"),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"message": "Download report error. Invalid request",
		}),
	)

	serviceAmounts := as.generateRandomReport(10, 5, 20, 20000)

	Test(as.T(),
		Post(getReportLinkPath),
		Send().Body().JSON(map[string]interface{}{
			"month": month,
			"year":  year,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Equal(map[string]interface{}{
			"link": link,
		}),
	)

	Test(as.T(),
		Get(link),
		Expect().Status().Equal(http.StatusOK),
		Expect().Headers("Content-Disposition").Equal(fmt.Sprintf("attachment; filename=%s", reportFilename)),
		Expect().Body().Bytes().Equal(as.serviceAmountsToCSV(serviceAmounts)),
	)
}
