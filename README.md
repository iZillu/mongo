# Try to work with MongoDB, It's gonna be funny! 
Let's imagine, that you have dosen users and the same count of groups. 
From time to time, these users can join any group and visa versa leave one.
So, It would be perfect, to have some methods, which will display the result of all of these moves in your DB.
Lucky for you, I got some ;)

## Requirements

Go 1.8 or above. 
Installed and runned MongoDB 

## Installation

Run the following command to get the sources:

```
git clone https://github.com/night-codes/mongo-prologue
cd mongo-prologue
```

Start the application:

```
go run main.go
```

Now the application running on address [http://localhost:8080/](http://localhost:8080/). 

## Methods

**User** have following parameters: 
*  login    : user's login
*  keyWord  : secrete word for recover password
*  age      : user's age
*  single   : marital status

```
| TYPE |        PATH        |         Description         |        Input value        |      Return value      |

  POST   /user/create         Create user                    User's parameters           Status / Error
  GET    /user/takeId         Find user by login             User's login                User's parameters
  GET    /user/get            Get user's info by ID          User's ID                   User's parameters
  POST   /user/update/<id>    Get user's info by ID          User's ID                   User's parameters
  GET    /user/delete         Delete user                    User's ID                   Status / Error
```

**Group** have following parameters: 
 * title    : group's title
 * minAge   : minimum age of user to become a member
 * members  : array of member's ID
 
 ```
| TYPE |        PATH        |         Description         |        Input value        |      Return value      |

  POST   /group/create         Create group                  Group's parameters          Status / Error
  GET    /group/takeId         Find group by title           Group's title               Group's parameters
  GET    /group/addUser        Add member to group           Group's ID, User's ID       Group's parameters
  GET    /group/getUsers       Get group's info by ID        Group's ID                  Array of members
  GET    /group/deleteUser     Delete member from group      Group's ID, User's ID       Group's parameters
  GET    /group/plus           Attach one group to another   GroupDst ID, GroupSrc ID    Status of each member
  GET    /group/minus          Kick one group from another   GroupDst ID, GroupSrc ID    Status / Error
  GET    /group/delete         Delete group                  Group's ID                  Status / Error
```  

### Good luck with my code ;)