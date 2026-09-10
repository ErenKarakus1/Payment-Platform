package services

import (
	"fmt"
	"log"

	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/mail"
	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/models"
)

const (
	EventPaymentCreated    = "payment.created"
	EventPaymentProcessing = "payment.processing"
	EventPaymentSucceeded  = "payment.succeeded"
	EventPaymentFailed     = "payment.failed"
	EventPaymentRefunded   = "payment.refunded"
)

func createMailBody(event models.PaymentEvent) string {
	amountFull := event.AmountCents / 100
	amountRemainder := event.AmountCents % 100
	switch event.EventType {

	case EventPaymentCreated:
		return fmt.Sprintf(`
Hi %s,

Your payment has been created successfully.

Payment ID: %s
Amount: %d.%02d %s
Status: Pending

We will notify you when the payment status changes.

Best regards,
Notification Service
		`, event.CustomerName, event.PaymentID.String(), amountFull, amountRemainder, event.Currency)

	case EventPaymentProcessing:
		return fmt.Sprintf(`
Hi %s,

Your payment is currently being processed.

Payment ID: %s
Amount: %d.%02d %s
Status: Processing

Best regards,
Notification Service
		`, event.CustomerName, event.PaymentID.String(), amountFull, amountRemainder, event.Currency)

	case EventPaymentSucceeded:
		return fmt.Sprintf(`
Hi %s,

Your payment has been successfully processed.

Payment ID: %s
Amount: %d.%02d %s
Status: Succeeded

Thank you for using our payment platform.

Best regards,
Notification Service
		`, event.CustomerName, event.PaymentID.String(), amountFull, amountRemainder, event.Currency)

	case EventPaymentFailed:
		return fmt.Sprintf(`
Hi %s,

Unfortunately, your payment could not be processed.

Payment ID: %s
Amount: %d.%02d %s
Status: Failed

Please try again or contact support if the problem persists.

Best regards,
Notification Service
		`, event.CustomerName, event.PaymentID.String(), amountFull, amountRemainder, event.Currency)

	case EventPaymentRefunded:
		return fmt.Sprintf(`
Hi %s,

A refund has been successfully processed for your payment.

Payment ID: %s
Refund Amount: %d.%02d %s
Status: Refunded

Best regards,
Notification Service
		`, event.CustomerName, event.PaymentID.String(), amountFull, amountRemainder, event.Currency)

	default:
		return ""
	}

}

func HandlePaymentEvent(event models.PaymentEvent, sender *mail.Sender) {
	body := createMailBody(event)
	switch event.EventType {
	case EventPaymentCreated:
		log.Printf(
			"Notification to %s: payment created for %d %s",
			event.CustomerEmail,
			event.AmountCents,
			event.Currency,
		)
		if err := sender.Send(event.CustomerEmail, "Payment Created", body); err != nil {
			log.Printf("Couldnt send message: %v, event: %+v", err, event)
		}
	case EventPaymentProcessing:
		log.Printf(
			"Notification to %s: payment is processing",
			event.CustomerEmail,
		)
		if err := sender.Send(event.CustomerEmail, "Payment Processing", body); err != nil {
			log.Printf("Couldnt send message: %v, event: %+v", err, event)
		}
	case EventPaymentSucceeded:
		log.Printf(
			"Notification to %s: payment succeeded",
			event.CustomerEmail,
		)
		if err := sender.Send(event.CustomerEmail, "Payment Succeeded", body); err != nil {
			log.Printf("Couldnt send message: %v, event: %+v", err, event)
		}
	case EventPaymentFailed:
		log.Printf(
			"Notification to %s: payment failed",
			event.CustomerEmail,
		)
		if err := sender.Send(event.CustomerEmail, "Payment Failed", body); err != nil {
			log.Printf("Couldnt send message: %v, event: %+v", err, event)
		}
	case EventPaymentRefunded:
		log.Printf(
			"Notification to %s: %d %s refunded",
			event.CustomerEmail,
			event.AmountCents,
			event.Currency,
		)
		if err := sender.Send(event.CustomerEmail, "Payment Refunded", body); err != nil {
			log.Printf("Couldnt send message: %v, event: %+v", err, event)
		}
	default:
		log.Printf("unknown payment event: %s", fmt.Sprint(event.EventType))
	}
}
