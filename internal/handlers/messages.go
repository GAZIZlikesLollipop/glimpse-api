package handlers

import (
	"api/hub"
	"api/internal"
	"api/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func GetChatMessages(c *gin.Context) {
	var messages []internal.Message
	rawUserId, exists := c.Get("userId")
	if !exists {
		log.Println("Ошибка получения данных с токена")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных с токена"})
		return
	}

	userId, ok := rawUserId.(int64)
	if !ok {
		log.Println("Ошибка преобрзаования айди")
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Ошибка преобразования айди"})
		return
	}
	rawReceiverId := c.Param("receiverId")
	receiverId, err := strconv.ParseInt(rawReceiverId, 10, 64)
	if err != nil {
		log.Println("Ошибка конвретации айди получателя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конвретации айди получателя"})
		return
	}
	if err := utils.MsgDb.Where("sender_id = ? AND receiver_id = ? OR sender_id = ? AND receiver_id = ?", userId, receiverId, receiverId, userId).Find(&messages).Error; err != nil {
		log.Println("Ошибка получения сообщений: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения сообщения"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func AddMessage(c *gin.Context) {
	rawUserId, exists := c.Get("userId")
	if !exists {
		log.Println("Ошибка получения данных с токена")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных с токена"})
		return
	}

	userId, ok := rawUserId.(int64)
	if !ok {
		log.Println("Ошибка преобрзаования айди")
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Ошибка преобразования айди"})
		return
	}
	rawReceiverId := c.Param("receiverId")
	receiverId, err := strconv.ParseInt(rawReceiverId, 10, 64)
	if err != nil {
		log.Println("Ошибка конвретации айди получателя: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конвретации айди получателя"})
		return
	}
	var sendMessage internal.Message
	if err := c.ShouldBindJSON(&sendMessage); err != nil {
		log.Println("Ошибка обработки тела: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки тела"})
		return
	}
	sendMessage.SenderId = &userId
	sendMessage.ReceiverId = &receiverId

	if err := utils.MsgDb.Create(&sendMessage).Error; err != nil {
		log.Println("Ошибка добавления сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления сообщения"})
		return
	}

	var messages []internal.Message
	if err := utils.MsgDb.Find(&messages).Error; err != nil {
		log.Println("Ошибка получения сообщений: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения сообщений"})
		return
	}

	c.JSON(http.StatusCreated, messages[len(messages)-1])
}

func DeleteMessage(c *gin.Context) {
	id := c.Param("id")

	var message internal.Message
	if err := utils.MsgDb.First(&message, id).Error; err != nil {
		log.Println("Ошибка получения сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения сообщения"})
		return
	}

	if err := utils.MsgDb.Delete(&message).Error; err != nil {
		log.Println("Ошибка Удаления сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка Удаления сообщения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успшеное удаление сообщения"})
}

func UpdateMessage(c *gin.Context) {
	var updateSentMessage, sentMessage internal.Message
	id := c.Param("id")

	if err := c.ShouldBindJSON(&updateSentMessage); err != nil {
		log.Println("Ошибка чтения тела запроса: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения тела запроса"})
		return
	}

	if err := utils.MsgDb.Where("id = ?", id).First(&sentMessage).Error; err != nil {
		log.Println("Ошибка получения сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получния сообщения"})
		return
	}

	if strings.TrimSpace(updateSentMessage.Content) != "" {
		sentMessage.Content = updateSentMessage.Content
	}

	if err := utils.MsgDb.Save(&sentMessage).Error; err != nil {
		log.Println("Ошибка обнволения сообщения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления сообщения"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Успешное обновление сообщения"})
}

var Hub *hub.Hub = hub.InitHub()

func WebSocket(c *gin.Context) {
	rawUserId, exists := c.Get("userId")
	if !exists {
		log.Println("Ошибка получения данных с токена")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных с токена"})
		return
	}

	userId, ok := rawUserId.(int64)
	if !ok {
		log.Println("Ошибка преобрзаования айди")
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "Ошибка преобразования айди"})
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	cnn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Ошибка создания webSocket соединения: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания websocket соединения"})
		return
	}
	defer cnn.Close()
	defer Hub.RemoveConnection(userId)

	var (
		user    internal.User
		msgs    []internal.Message
		request internal.WebSocketResponse
	)
	userResp, err := getUserDataRequest(userId)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения пользователя"))
		return
	}

	user = *userResp

	Hub.AddConnection(user.Id, cnn)

	cnn.WriteMessage(websocket.TextMessage, []byte("WebSocket соединение успешно создано"))

	for {
		_, strId, err := cnn.ReadMessage()
		if err != nil {
			log.Println("Ошибка получения сообщения: ", err)
			cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получения сообщения"))
			break
		}

		userResp, err := getUserDataRequest(userId)
		if err != nil {
			log.Println("Ошибка получения пользователя: ", err)
			cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения пользователя"))
			return
		}

		user = *userResp

		receiverId, recErr := strconv.ParseInt(string(strId), 10, 0)
		if recErr != nil {
			request = internal.WebSocketResponse{
				User:     user,
				Messages: []internal.Message{},
			}
		} else {
			if err := utils.MsgDb.Where("sender_id = ? AND receiver_id = ? OR sender_id = ? AND receiver_id = ?", userId, receiverId, receiverId, userId).Find(&msgs).Error; err != nil {
				log.Println("Ошибка получения сообщений: ", err)
				cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получения сообщений"))
				return
			}

			request = internal.WebSocketResponse{
				User:     user,
				Messages: msgs,
			}
		}

		if data, err := json.Marshal(&request); err != nil {
			log.Println("Ошибка преобразования типов: ", err)
			cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка преобразования типов"))
			break
		} else {
			if recErr == nil {
				if friendCnn, ok := Hub.GetConnection(receiverId); ok {
					friend, err := getUserDataRequest(receiverId)
					if err != nil {
						log.Println("Ошибка получения друга: ", err)
						cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения друга"))
						return
					}

					data := internal.WebSocketResponse{
						User:     *friend,
						Messages: msgs,
					}
					friendData, err := json.Marshal(data)
					if err != nil {
						log.Println("Ошибка кодирования json: ", err)
						cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка кодирования json"))
						return
					}
					friendCnn.WriteMessage(websocket.TextMessage, friendData)
				}
			} else {
				for _, userFriend := range user.Friends {
					if cnn, ok := Hub.GetConnection(userFriend.Id); ok {
						friend, err := getUserDataRequest(userFriend.Id)
						if err != nil {
							log.Println("Ошибка получения друга: ", err)
							cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка получеения друга"))
							return
						}
						data := internal.WebSocketResponse{
							User:     *friend,
							Messages: []internal.Message{},
						}
						friendData, err := json.Marshal(data)
						if err != nil {
							log.Println("Ошибка кодирования json: ", err)
							cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка кодирования json"))
							return
						}
						cnn.WriteMessage(websocket.TextMessage, friendData)
					}
				}
			}
			if err := cnn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("Ошибка отправки json: ", err)
				cnn.WriteMessage(websocket.TextMessage, []byte("Ошибка отпрваки json"))
				return
			}
		}
	}
	cnn.WriteMessage(websocket.TextMessage, []byte("WebSocket соединение успешно разорвано!"))
}
