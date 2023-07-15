# go-linebot-service

The `linebotservice` package provides a Line bot service that handles Line bot events by processing messages using a message service.

## Installation

You can install the `linebotservice` package using `go get`:

```
go get github.com/Zhima-Mochi/go-linebot-service/linebotservice
```

## Usage

To use the `linebotservice` package, you need to create a `LineBotService` instance with a Line bot client and a message service:

```go
import (
	"github.com/Zhima-Mochi/go-linebot-service/messageservice"
	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory/chatgpt"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sashabaranov/go-openai"
)

// Create a Line bot client
linebotClient, err := linebot.New("YOUR_CHANNEL_SECRET", "YOUR_CHANNEL_ACCESS_TOKEN")
if err != nil {
    log.Fatal(err)
}

// Create a message service
// Default message core is echo
messageService := messageservice.NewMessageService()

// Set another default message core
openaiClient := openai.NewClient("sk-xxx")
messageService.SetDefaultMessageCore(chatgpt.NewMessageCore(openaiClient, linebotClient))

// Create a LineBotService instance
bot := linebotservice.NewLineBotService(client, messageService)
```

You can then use the `HandleEvents` method of the `LineBotService` instance to handle incoming HTTP requests:

```go
http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
    bot.HandleEvents(w, req)
})
```

The `HandleEvents` method parses the Line bot events from the HTTP request, processes messages using the message service, and replies to the events with the processed messages.