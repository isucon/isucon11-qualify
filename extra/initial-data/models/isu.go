package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/isucon/isucon11-qualify/bench/random"
)

type Isu struct {
	User        User
	JIAIsuUUID  string
	Name        string
	Image       []byte
	Character   string
	CharacterId int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewIsu(user User) Isu {
	u, _ := uuid.NewRandom()
	createdAt := random.TimeAfterArg(user.CreatedAt)

	image := defaultImage()
	character, characterID := random.CharacterWithID()

	return Isu{
		user,
		u.String(),
		random.IsuName(),
		image,
		character,
		characterID,
		createdAt,
		createdAt,
	}
}

func NewIsuWithCreatedAt(user User, createdAt time.Time) Isu {
	u, _ := uuid.NewRandom()
	image := defaultImage()
	character, characterID := random.CharacterWithID()

	return Isu{
		user,
		u.String(),
		random.IsuName(),
		image,
		character,
		characterID,
		createdAt,
		createdAt,
	}
}

func NewIsuWithCharacterId(user User, characterID int) Isu {
	u, _ := uuid.NewRandom()
	createdAt := random.TimeAfterArg(user.CreatedAt)

	image := defaultImage()
	character := random.CharacterData[characterID%len(random.CharacterData)]

	return Isu{
		user,
		u.String(),
		random.IsuName(),
		image,
		character,
		characterID,
		createdAt,
		createdAt,
	}
}
func defaultImage() []byte {
	bytes, err := ioutil.ReadFile(defaultImagePath)
	if err != nil {
		log.Fatalf("%+v", fmt.Errorf("%w", err))
	}
	return bytes
}

func (i *Isu) WithUpdateName() error {
	i.Name = random.IsuName()
	i.UpdatedAt = random.TimeAfterArg(i.UpdatedAt)
	return nil
}
func (i *Isu) WithUpdateImage() error {
	var err error
	i.Image, err = random.Image()
	i.UpdatedAt = random.TimeAfterArg(i.UpdatedAt)
	return err
}

func (i Isu) Create() error {
	if _, err := db.Exec("INSERT INTO isu(`jia_isu_uuid`,`name`,`image`,`character`,`jia_user_id`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?)",
		i.JIAIsuUUID, i.Name, i.Image, i.Character, i.User.JIAUserID,
		i.CreatedAt, i.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert isu: %w", err)
	}
	return nil
}
