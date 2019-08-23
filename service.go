/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   service.go                                         :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: hmuravch <neji926629@gmail.com>            +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2019/08/15 16:31:38 by hmuravch          #+#    #+#             */
/*   Updated: 2019/08/23 17:22:05 by hmuravch         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	ai "github.com/night-codes/mgo-ai"
	"github.com/night-codes/tokay"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
)

// Service is the main class with all methods to work with DB
type Service struct {
	sync.Mutex
	userCashMap map[uint64]*userStruct
}

// obj ussually is used like field of db -> [field's name]: value
type obj map[string]interface{}

// dataBase is list of collections
type dataBase struct {
	UserCol    *mgo.Collection
	GroupCol   *mgo.Collection
	RequestCol *mgo.Collection
}

/**
*** SITE STRUCTS
**/

type groupParams struct {
	ID     uint64 `form:"-" bson:"_id" json:"id"`
	Title  string `form:"title" bson:"title" json:"title" valid:"required, min(3), max(16)" `
	MinAge uint   `form:"minAge" bson:"minAge" json:"minAge" valid:"required" `
}

type infoGroup struct {
	ID    uint64 `form:"-" bson:"_id" json:"id"`
	Title string `form:"title" bson:"title" json:"title" valid:"required, min(3), max(16)" `
}

type signInParams struct {
	ID    uint64 `form:"-" bson:"_id" json:"_id"`
	Login string `form:"login" bson:"login" json:"login" valid:"required,min(3),max(40)"`
	Hash  string `form:"hash" bson:"hash" json:"hash" valid:"required"`
}

type userStruct struct {
	ID               uint64      `form:"-" bson:"_id" json:"_id"`
	Hash             string      `form:"hash" bson:"hash" json:"hash" valid:"required"`
	Login            string      `form:"login" bson:"login" json:"login" valid:"required,min(3),max(40)"`
	Email            string      `form:"email" bson:"email" json:"email" valid:"required,min(3),max(42)"`
	RegistrationTime string      `form:"time" bson:"time" json:"time"`
	Status           bool        `form:"status" bson:"status" json:"status"`
	IDGroups         []uint64    `form:"groups" bson:"groups" json:"-"`
	Groups           []infoGroup `form:"-" bson:"-" json:"groups"`

	Stored      bool      `form:"-" bson:"-"`
	ExpiredTime time.Time `form:"-" bson:"-"`
}

type requestStruct struct {
	ID        uint64 `form:"-" bson:"_id" json:"_id"`
	UserLogin string `form:"login" bson:"login" json:"login"`
	Request   string `form:"request" bson:"request" json:"request" valid:"required"`
	Status    bool   `form:"status" bson:"status" json:"status"`
}

type status struct {
	Status bool `form:"status" bson:"status" json:"status"`
}

/**
*** VARIABLES
**/

var (
	s   Service
	err error
	db  dataBase
)

/**
*** SERVICE METHODS
**/

func errorAlert(errMsg string, err error, c *tokay.Context) bool {
	if err != nil {
		c.JSON(400, obj{
			"err": errMsg,
		})
		return true
	}
	return false
}

func (s *Service) toStoreCash() {
	for {
		for key, user := range s.userCashMap {
			if user.Stored {
				if err := db.UserCol.UpdateId(key, obj{"set": user}); err != nil {
					log.Println("Cash-Worker toStore - goes down!")
				} else {
					s.Lock()
					s.userCashMap[key].Stored = false
					s.Unlock()
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Service) toFreeCash() {
	for {
		for _, user := range s.userCashMap {
			if user.Stored == false {
				curTime := time.Now()
				if curTime.Sub(user.ExpiredTime).Minutes() > 1 {
					delete(s.userCashMap, user.ID) // ???
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

/**
*** METHODS
**/

//	USER

// CreateUser insert user in Data Base | input: - | return: err / ["ok"]: "true"
func (s *Service) CreateUser(c *tokay.Context) {
	uParams := userStruct{}

	err = c.BindJSON(&uParams)
	if errorAlert("Bind fall down", err, c) {
		return
	}
	uParams.ID = ai.Next("user")

	hash, err := bcrypt.GenerateFromPassword([]byte(uParams.Hash), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	uParams.Hash = string(hash)

	time := time.Now().Format("Jan 2, 2006 at 3:04pm")
	uParams.RegistrationTime = time

	err = db.UserCol.Insert(uParams)
	if errorAlert("Error: Input parameteres already used", err, c) {
		return
	}

	c.JSON(200, obj{"ok": "true"})
}

// GetAllUsers returns all users  | input: - | return: err / []userStruct{}
func (s *Service) GetAllUsers(c *tokay.Context) {
	allUsers := []userStruct{}

	err := db.UserCol.Find(obj{}).Sort("-time").All(&allUsers)
	if errorAlert("Method GetAll goes down...", err, c) {
		return
	}

	c.JSON(200, obj{"users": allUsers})
}

// countUsers returns all users  | input: - | return: err / []userStruct{}
func (s *Service) countUsers(c *tokay.Context) {
	count, err := db.UserCol.Find(obj{}).Count()
	if errorAlert("Method count goes down...", err, c) {
		return
	}

	c.JSON(200, obj{"count": count})
}

// GetNUsers returns all users  | input: - | return: err / []userStruct{}
func (s *Service) GetNUsers(c *tokay.Context) {
	nUsers := []userStruct{}
	skip, _ := strconv.ParseInt(c.Param("skip"), 10, 64)
	limit, _ := strconv.ParseInt(c.Param("limit"), 10, 64)

	err := db.UserCol.Find(obj{}).Skip(int(skip)).Limit(int(limit)).Sort("-time").All(&nUsers)
	if errorAlert("Method GetAll goes down...", err, c) {
		return
	}

	c.JSON(200, obj{"users": nUsers})
}

// GetUserByID returns user by ID | input: <id> | return: err / userStruct{}
func (s *Service) GetUserByID(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	findUser := userStruct{}
	err := db.UserCol.FindId(ID).One(&findUser)
	if errorAlert("User can not be found", err, c) {
		return
	}

	listGroups := []infoGroup{}

	db.GroupCol.Find(
		obj{
			"_id": obj{
				"$in": findUser.IDGroups,
			},
		},
	).All(&listGroups)
	findUser.Groups = listGroups
	c.JSON(200, findUser)
}

// GetUserByLogin returns user by login | input: <login> | return: err / userStruct{}
func (s *Service) GetUserByLogin(c *tokay.Context) {
	findUser := userStruct{}
	uLogin := c.Param("login")
	filter := obj{"login": uLogin}

	err = db.UserCol.Find(filter).One(&findUser)
	if errorAlert("User can not be found", err, c) {
		return
	}
	c.JSON(200, findUser)
}

// UpdateUserByID updates user by ID | input: <id> | return: err / ["ok"]: "true"
func (s *Service) UpdateUserByID(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	findUser := userStruct{}
	err = db.UserCol.FindId(ID).One(obj{})
	if errorAlert("User can not be found", err, c) {
		return
	}
	err = c.BindJSON(&findUser)
	if errorAlert("Invalid parameter", err, c) {
		return
	}
	filter := obj{"$set": findUser}
	err = db.UserCol.UpdateId(ID, filter)
	if errorAlert("User was not updated", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// UpdateUserStatusByLogin updates user by login | input: <login> | return: err / ["ok"]: "true"
func (s *Service) UpdateUserStatusByLogin(c *tokay.Context) {
	uLogin := c.Param("login")
	findUser := status{}

	err = db.UserCol.Find(obj{"login": uLogin}).One(obj{})
	if errorAlert("User can not be found", err, c) {
		return
	}
	err = c.BindJSON(&findUser)
	if errorAlert("Invalid parameter", err, c) {
		return
	}
	filter := obj{"$set": findUser}
	err = db.UserCol.Update(obj{"login": uLogin}, filter)
	if errorAlert("User was not updated", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// DeleteUser deletes user by ID | input: <id> | return: err / ["ok"]: "true"
func (s *Service) DeleteUser(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	filter := obj{"_id": ID}
	err = db.UserCol.Remove(filter)
	if errorAlert("User was not deleted", err, c) {
		return
	}

	c.JSON(200, obj{"ok": "true"})
}

// REQUEST

// CreateRequest insert user in Data Base | input: - | return: err / ["ok"]: "true"
func (s *Service) CreateRequest(c *tokay.Context) {
	request := requestStruct{}

	err = c.BindJSON(&request)
	if errorAlert("Bind fall down", err, c) {
		return
	}
	request.ID = ai.Next("users")
	request.Status = false

	err = db.RequestCol.Insert(request)
	if errorAlert("Error: Request was not added to DB", err, c) {
		return
	}

	c.JSON(200, obj{"ok": "true"})
}

// GetRequestByID returns request by ID | input: <id> | return: err / requestStruct{}
func (s *Service) GetRequestByID(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	findRequest := requestStruct{}
	err := db.RequestCol.FindId(ID).One(&findRequest)
	if errorAlert("Request can not be found", err, c) {
		return
	}

	c.JSON(200, findRequest)
}

// GetAllRequests returns all requests  | input: - | return: err / []requestStruct{}
func (s *Service) GetAllRequests(c *tokay.Context) {
	allRequests := []requestStruct{}

	err := db.RequestCol.Find(obj{}).Sort("-login").All(&allRequests)
	if errorAlert("Method GetAll goes down...", err, c) {
		return
	}
	c.JSON(200, obj{"requests": allRequests})
}

// UpdateRequestByID updates request by ID | input: <id> | return: err / ["ok"]: "true"
func (s *Service) UpdateRequestByID(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	err = db.RequestCol.FindId(ID).One(obj{})
	if errorAlert("Request can not be found", err, c) {
		return
	}
	err = db.RequestCol.UpdateId(ID, obj{"$set": obj{"status": true}})
	if errorAlert("Request was not updated", err, c) {
		return
	}
	c.JSON(200, obj{"ok": true})
}

// 	GROUP

// CreateGroup insert group in Data Base | input: - | return: map[string]interface{}
func (s *Service) CreateGroup(c *tokay.Context) {
	gParam := groupParams{}

	err = c.Bind(&gParam)
	if errorAlert("Bind fall down", err, c) {
		return
	}
	gParam.ID = ai.Next("group")

	err = db.GroupCol.Insert(gParam)
	if errorAlert("Group was not created", err, c) {
		return
	}
	c.JSON(200, obj{
		"ok": "true",
	})
}

// GetGroupByID returns group by login | input: <id> | return: err / groupParams{}
func (s *Service) GetGroupByID(c *tokay.Context) {
	findGroup := groupParams{}
	ID := uint64(c.ParamUint("id"))

	err := db.GroupCol.FindId(ID).One(&findGroup)
	if errorAlert("Group can not be found", err, c) {
		return
	}
	c.JSON(200, findGroup)
	return
}

// GetGroupByTitle returns group by title | input: <title> | return: err / userStruct{}
func (s *Service) GetGroupByTitle(c *tokay.Context) {
	title := c.Param("title")
	findGroup := groupParams{}

	filter := obj{"title": title}
	err = db.GroupCol.Find(filter).One(&findGroup)
	if errorAlert("Group can not be found", err, c) {
		return
	}
	c.JSON(200, findGroup)
}

// UpdateGroupByID updates group by ID | input: <id> | return: err / ["ok"]: "true"
func (s *Service) UpdateGroupByID(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	curGroup := groupParams{}
	err = db.GroupCol.FindId(ID).One(&curGroup)
	if errorAlert("Group can not be found", err, c) {
		return
	}
	err = c.BindJSON(&curGroup)
	if errorAlert("Invalid parameter", err, c) {
		return
	}
	filter := obj{"$set": curGroup}
	err = db.GroupCol.UpdateId(ID, filter)
	if errorAlert("Group was not updated", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// DeleteGroup deletes Group by ID | input: <id> | return: err / ["ok"]: "true"
func (s *Service) DeleteGroup(c *tokay.Context) {
	ID := uint64(c.ParamUint("id"))

	filter := obj{"_id": ID}
	err = db.GroupCol.Remove(filter)
	if errorAlert("Group was not deleted", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// AddUserToGroup adds user to group | input: <groupID>, <userID> | return: err / ["ok"]: "true"
func (s *Service) AddUserToGroup(c *tokay.Context) {
	gID := uint64(c.QueryUint("groupID"))
	uID := uint64(c.QueryUint("userID"))

	curGroup := groupParams{}
	curUser := userStruct{}

	err = db.GroupCol.Find(obj{"_id": gID}).One(&curGroup)
	if errorAlert("Group does not exist", err, c) {
		return
	}
	err = db.UserCol.Find(obj{"_id": uID}).One(&curUser)
	if errorAlert("User does not exist", err, c) {
		return
	}
	err = db.UserCol.UpdateId(uID, obj{"$addToSet": obj{"groups": gID}})
	if errorAlert("Member was not added", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// DeleteUserFromGroup deletes user from group | input: <groupID>, <userID> | return: err / ["ok"]: "true"
func (s *Service) DeleteUserFromGroup(c *tokay.Context) {
	gID := uint64(c.QueryUint("groupID"))
	uID := uint64(c.QueryUint("userID"))

	curGroup := groupParams{}

	err = db.GroupCol.Find(obj{"_id": gID}).One(&curGroup)
	if errorAlert("Group does not exist", err, c) {
		return
	}
	err = db.UserCol.Find(obj{"_id": uID}).One(obj{})
	if errorAlert("User does not exist", err, c) {
		return
	}
	err = db.UserCol.UpdateId(uID, obj{
		"$pull": obj{
			"groups": gID,
		},
	})
	if errorAlert("Member was not deleted", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// GetAllMembersOfGroup add one group of users to another. | input: <groupID>| return: err / []userStruct{}
func (s *Service) GetAllMembersOfGroup(c *tokay.Context) {
	gID := uint64(c.QueryUint("groupID"))

	err = db.GroupCol.FindId(gID).One(obj{})
	if errorAlert("Group does not exist", err, c) {
		return
	}
	memberList := []userStruct{}
	err = db.UserCol.Find(obj{
		"groups": obj{
			"$in": []uint64{gID},
		},
	}).All(&memberList)
	if errorAlert("Syntax error", err, c) {
		return
	}
	if len(memberList) == 0 {
		c.JSON(400, obj{
			"Message": "Group is empty",
		})
	}
	c.JSON(200, memberList)
}

// PlusGroupToGroup add one group of users to another. | input: <ID1>, <ID2> | return: err / ["ok"]: "true"
func (s *Service) PlusGroupToGroup(c *tokay.Context) {
	dstID := uint64(c.QueryUint("ID1"))
	srcID := uint64(c.QueryUint("ID2"))

	dstGroup := groupParams{}

	err = db.GroupCol.Find(obj{"_id": dstID}).One(&dstGroup)
	if errorAlert("Destination group does not exist", err, c) {
		return
	}
	err = db.GroupCol.Find(obj{"_id": srcID}).One(obj{})
	if errorAlert("Source group does not exist", err, c) {
		return
	}
	_, err := db.UserCol.UpdateAll(
		obj{
			"groups": srcID,
			"age": obj{
				"$gte": dstGroup.MinAge,
			},
		}, obj{
			"$addToSet": obj{
				"groups": dstGroup.ID,
			},
		},
	)
	if errorAlert("Group  was not added", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

// MinusGroupFromGroup deletes one group of users from another. | input: <ID1>, <ID2> | return: err / ["ok"]: "true"
func (s *Service) MinusGroupFromGroup(c *tokay.Context) {
	dstID := uint64(c.QueryUint("ID1"))
	srcID := uint64(c.QueryUint("ID2"))

	dstGroup := groupParams{}
	srcGroup := groupParams{}

	err = db.GroupCol.Find(obj{"_id": dstID}).One(&dstGroup)
	if errorAlert("Destination group does not exist", err, c) {
		return
	}
	err = db.GroupCol.Find(obj{"_id": srcID}).One(&srcGroup)
	if errorAlert("Source group does not exist", err, c) {
		return
	}
	_, err = db.UserCol.UpdateAll(
		obj{
			"groups": srcID,
		}, obj{
			"$pull": obj{
				"groups": dstGroup.ID,
			},
		},
	)
	if errorAlert("Group  was not added", err, c) {
		return
	}
	c.JSON(200, obj{"ok": "true"})
}

/**
*** INIT
**/

func init() {

	session, err := mgo.Dial(":27017")
	if err != nil {
		panic(err)
	}

	db = dataBase{
		UserCol:    session.DB("site").C("user"),
		GroupCol:   session.DB("site").C("group"),
		RequestCol: session.DB("site").C("request"),
	}

	var email, login bool

	if listID, err := db.UserCol.Indexes(); err != nil {
		fmt.Println(err.Error())
	} else {
		for _, v := range listID {
			if len(v.Key) == 1 {
				if v.Key[0] == "email" {
					email = true
				} else if v.Key[0] == "login" {
					login = true
				}
			}
		}
	}
	if !email {
		db.UserCol.EnsureIndex(mgo.Index{Key: []string{"email"}, Unique: true})
	}
	if !login {
		db.UserCol.EnsureIndex(mgo.Index{Key: []string{"login"}, Unique: true})
	}

	ai.Connect(session.DB("site").C("counts"))

	go s.toStoreCash()
	go s.toFreeCash()
}

/**
*** MAIN
**/

func main() {
	router := tokay.New()

	// USER

	userRoute := router.Group("/user")

	userRoute.POST("/create", func(c *tokay.Context) {
		s.CreateUser(c)
	})

	userRoute.GET("/getAll", func(c *tokay.Context) {
		s.GetAllUsers(c)
	})

	userRoute.GET("/count", func(c *tokay.Context) {
		s.countUsers(c)
	})

	userRoute.GET("/getN/<skip>/<limit>", func(c *tokay.Context) {
		s.GetNUsers(c)
	})

	userRoute.GET("/get/<id>", func(c *tokay.Context) {
		s.GetUserByID(c)
	})

	userRoute.GET("/getByLogin/<login>", func(c *tokay.Context) {
		s.GetUserByLogin(c)
	})

	userRoute.POST("/update/<id>", func(c *tokay.Context) {
		s.UpdateUserByID(c)
	})

	userRoute.POST("/updateStatusByLogin/<login>", func(c *tokay.Context) {
		s.UpdateUserStatusByLogin(c)
	})

	userRoute.DELETE("/delete/<id>", func(c *tokay.Context) {
		s.DeleteUser(c)
	})

	// REQUEST

	requestRoute := router.Group("/request")

	requestRoute.POST("/create", func(c *tokay.Context) {
		s.CreateRequest(c)
	})

	requestRoute.GET("/getAll", func(c *tokay.Context) {
		s.GetAllRequests(c)
	})

	requestRoute.GET("/get/<id>", func(c *tokay.Context) {
		s.GetRequestByID(c)
	})

	requestRoute.POST("/update/<id>", func(c *tokay.Context) {
		s.UpdateRequestByID(c)
	})

	// GROUP

	groupRoute := router.Group("/group")

	groupRoute.POST("/create", func(c *tokay.Context) {
		s.CreateGroup(c)
	})

	groupRoute.GET("/get/<id>", func(c *tokay.Context) {
		s.GetGroupByID(c)
	})

	groupRoute.GET("/getByTitle/<title>", func(c *tokay.Context) {
		s.GetGroupByTitle(c)
	})

	groupRoute.POST("/update/<id>", func(c *tokay.Context) {
		s.UpdateGroupByID(c)
	})

	groupRoute.DELETE("/delete/<id>", func(c *tokay.Context) {
		s.DeleteGroup(c)
	})

	groupRoute.GET("/addUser", func(c *tokay.Context) {
		s.AddUserToGroup(c)
	})

	groupRoute.GET("/deleteUser", func(c *tokay.Context) {
		s.DeleteUserFromGroup(c)
	})

	groupRoute.GET("/getUsers", func(c *tokay.Context) {
		s.GetAllMembersOfGroup(c)
	})

	groupRoute.GET("/plus", func(c *tokay.Context) {
		s.PlusGroupToGroup(c)
	})

	groupRoute.GET("/minus", func(c *tokay.Context) {
		s.MinusGroupFromGroup(c)
	})

	router.Run(":4000", "Application started at http://localhost%s")
}
