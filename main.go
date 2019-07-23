package main

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/night-codes/tokay"
	"gopkg.in/mgo.v2"
)

/**
*** STRUCTS
**/

type obj map[string]interface{}

type userParam struct {
	Login   string `form:"login" bson:"login" valid:"min(3), max(16)"`
	KeyWord string `form:"keyWord" bson:"keyWord" valid:"min(1), max(99)"`
	Age     uint   `form:"age" bson:"age"`
	Single  bool   `form:"single" bson:"single"`
}

type groupParam struct {
	Title   string          `form:"title" bson:"title" valid:"requird, min(3), max(16)"`
	MinAge  uint            `form:"minAge" bson:"minAge" valid:"requird"`
	Members []bson.ObjectId `form:"members" bson:"members" valid:"requird"`
}

type dataBase struct {
	UserCol  *mgo.Collection
	GroupCol *mgo.Collection
}

/**
*** SECONDARY METHODS
**/

func errorAlert(errMsg string, err error, c *tokay.Context) bool {
	if err != nil {
		c.JSON(400, obj{
			"Error": errMsg,
		})
		return true
	}
	return false
}

/**
*** DATA BASE METHODS
**/

func (db *dataBase) userGetInfo(c *tokay.Context, id bson.ObjectId) (userParam, error) {
	filter := obj{"_id": id}
	findUser := userParam{}
	err := db.UserCol.Find(filter).One(&findUser)
	if errorAlert("User can not be found", err, c) {
		return findUser, err
	}
	return findUser, nil
}

/**
*** MAIN
**/

func main() {
	api := tokay.New()

	session, err := mgo.Dial(":27017")
	if err != nil {
		panic(err)
	}

	var db = dataBase{
		UserCol:  session.DB("db").C("user"),
		GroupCol: session.DB("db").C("group"),
	}

	api.POST("/user/create", func(c *tokay.Context) {
		uParam := userParam{}

		err = c.Bind(&uParam)
		if errorAlert("Bind fall down", err, c) {
			return
		}
		err = db.UserCol.Insert(uParam)
		if errorAlert("User was not created", err, c) {
			return
		}
		c.JSON(200, obj{"Status": "ok"})
	})

	api.GET("/user/takeId", func(c *tokay.Context) {
		filter := obj{"login": c.Query("login")}
		findUser := userParam{}

		err = db.UserCol.Find(filter).One(&findUser)
		if errorAlert("User can not be found", err, c) {
			return
		}
		c.JSON(200, findUser)
	})

	api.GET("/user/get", func(c *tokay.Context) {
		inputID := c.Query("id")

		if bson.IsObjectIdHex(inputID) != false {
			id := bson.ObjectIdHex(inputID)
			findUser, err := db.userGetInfo(c, id)
			if err != nil {
				return
			}
			c.JSON(200, findUser)
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.POST("/user/update/<id>", func(c *tokay.Context) {
		input := c.Param("id")

		if bson.IsObjectIdHex(input) != false {
			id := bson.ObjectIdHex(input)
			findUser := userParam{}
			err = db.UserCol.FindId(id).One(&findUser)
			if errorAlert("User can not be found", err, c) {
				return
			}
			err = c.Bind(&findUser)
			if errorAlert("Invalid parameter", err, c) {
				return
			}
			filter := obj{"$set": findUser}
			err = db.UserCol.UpdateId(id, filter)
			if errorAlert("User was not updated", err, c) {
				return
			}
			c.JSON(200, obj{"Updated": "true"})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.GET("/user/delete", func(c *tokay.Context) {
		id := c.Query("id")

		if bson.IsObjectIdHex(id) != false {
			tmp := bson.ObjectIdHex(id)
			filter := obj{"_id": tmp}
			err = db.UserCol.Remove(filter)
			if errorAlert("User was not deleted", err, c) {
				return
			}
			c.JSON(200, obj{
				"Status": "ok",
			})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.POST("/group/create", func(c *tokay.Context) {
		gParam := groupParam{}

		err = c.Bind(&gParam)
		if errorAlert("Bind fall down", err, c) {
			return
		}
		err = db.GroupCol.Insert(gParam)
		if errorAlert("Group was not created", err, c) {
			return
		}
		c.JSON(200, obj{
			"Result": "Group was created",
		})
	})

	api.GET("/group/takeId", func(c *tokay.Context) {
		filter := obj{"title": c.Query("title")}
		findGroup := groupParam{}

		err = db.GroupCol.Find(filter).One(&findGroup)
		if errorAlert("Group can not be found", err, c) {
			return
		}
		c.JSON(200, findGroup)
	})

	api.GET("/group/addUser", func(c *tokay.Context) {
		idOne := c.Query("id")
		idTwo := c.Query("userid")

		if bson.IsObjectIdHex(idOne) && bson.IsObjectIdHex(idTwo) {
			curGroup := groupParam{}
			gID, uID := bson.ObjectIdHex(idOne), bson.ObjectIdHex(idTwo)

			err = db.GroupCol.Find(obj{"_id": gID}).One(&curGroup)
			if errorAlert("Group does not exist", err, c) {
				return
			}
			err = db.UserCol.Find(obj{"_id": uID}).One(obj{})
			if errorAlert("User does not exist", err, c) {
				return
			}
			err = db.GroupCol.UpdateId(gID, obj{"$push": obj{"members": uID}})
			if errorAlert("Member was not added", err, c) {
				return
			}
			c.JSON(200, obj{
				"Status": "ok",
			})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.GET("/group/getUsers", func(c *tokay.Context) {
		idOne := c.Query("id")

		if bson.IsObjectIdHex(idOne) {
			curGroup := groupParam{}
			memberList := []userParam{}
			gID := bson.ObjectIdHex(idOne)

			err = db.GroupCol.Find(obj{"_id": gID}).One(&curGroup)
			if errorAlert("Group does not exist", err, c) {
				return
			}
			err = db.GroupCol.FindId(gID).One(&curGroup)
			if errorAlert("Invalid usage", err, c) {
				return
			}
			if len(curGroup.Members) == 0 {
				c.JSON(404, obj{
					"Message": "Group is empty",
				})
				return
			}
			for i := range curGroup.Members {
				tmpUser, err := db.userGetInfo(c, curGroup.Members[i])
				if err != nil {
					return
				}
				memberList = append(memberList, tmpUser)
			}
			c.JSON(200, memberList)
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.GET("/group/deleteUser", func(c *tokay.Context) {
		idOne := c.Query("id")
		idTwo := c.Query("userid")

		if bson.IsObjectIdHex(idOne) && bson.IsObjectIdHex(idTwo) {
			curGroup := groupParam{}
			gID, uID := bson.ObjectIdHex(idOne), bson.ObjectIdHex(idTwo)

			err = db.GroupCol.Find(obj{"_id": gID}).One(&curGroup)
			if errorAlert("Group does not exist", err, c) {
				return
			}
			err = db.UserCol.Find(obj{"_id": uID}).One(obj{})
			if errorAlert("User does not exist", err, c) {
				return
			}
			err = db.GroupCol.UpdateId(gID, obj{"$pull": obj{"members": uID}})
			if errorAlert("Member was not deleted", err, c) {
				return
			}
			c.JSON(200, obj{
				"Status": "ok",
			})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.GET("/group/plus", func(c *tokay.Context) {
		idOne := c.Query("id1")
		idTwo := c.Query("id2")

		if bson.IsObjectIdHex(idOne) && bson.IsObjectIdHex(idTwo) {
			dstGr := groupParam{}
			srcGr := groupParam{}
			dstID, srcID := bson.ObjectIdHex(idOne), bson.ObjectIdHex(idTwo)

			err = db.GroupCol.Find(obj{"_id": dstID}).One(&dstGr)
			if errorAlert("Destination group does not exist", err, c) {
				return
			}
			err = db.GroupCol.Find(obj{"_id": srcID}).One(&srcGr)
			if errorAlert("Source group does not exist", err, c) {
				return
			}
			for i := range srcGr.Members {
				err = db.GroupCol.UpdateId(dstID, obj{"$addToSet": obj{"members": srcGr.Members[i]}})
				if errorAlert("Member was not added", err, c) {
					return
				}
			}
			c.JSON(200, obj{
				"Status": "ok",
			})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.GET("/group/minus", func(c *tokay.Context) {
		idOne := c.Query("id1")
		idTwo := c.Query("id2")

		if bson.IsObjectIdHex(idOne) && bson.IsObjectIdHex(idTwo) {
			dstGr := groupParam{}
			srcGr := groupParam{}
			dstID, srcID := bson.ObjectIdHex(idOne), bson.ObjectIdHex(idTwo)

			err = db.GroupCol.Find(obj{"_id": dstID}).One(&dstGr)
			if errorAlert("Destination group does not exist", err, c) {
				return
			}
			err = db.GroupCol.Find(obj{"_id": srcID}).One(&srcGr)
			if errorAlert("Source group does not exist", err, c) {
				return
			}
			err = db.GroupCol.UpdateId(dstID, obj{"$pullAll": obj{"members": srcGr.Members}})
			c.JSON(200, obj{
				"Status": "ok",
			})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.GET("/group/delete", func(c *tokay.Context) {
		id := c.Query("id")

		if bson.IsObjectIdHex(id) != false {
			tmp := bson.ObjectIdHex(id)
			filter := obj{"_id": tmp}
			err = db.GroupCol.Remove(filter)
			if errorAlert("Group was not deleted", err, c) {
				return
			}
			c.JSON(200, obj{
				"Status": "ok",
			})
			return
		}
		c.JSON(400, obj{
			"Error": "Invalid id",
		})
	})

	api.Run(":8080", "Application started at http://localhost%s")

}
