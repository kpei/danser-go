package play

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/thehowl/go-osuapi"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
)

type circle struct {
	p image.Point
	r int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.25, float64(y-c.p.Y)+0.25, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

type AvatarCircle struct {
	name      string
	teamColor color.Color
	avatar    *sprite.Sprite
}

func NewAvatarCircle(name string, teamColor color.Color) *AvatarCircle {
	avatar := &AvatarCircle{
		name:      name,
		teamColor: teamColor,
	}
	avatar.LoadAvatarUser(name)

	if avatar.avatar == nil {
		avatar.LoadDefaultAvatar()
	}

	return avatar
}

func (avatar *AvatarCircle) Draw(x float64, y float64, alpha float64, batch *batch.QuadBatch) {
	avatar.avatar.SetPosition(vector.NewVec2d(x, y))
	avatar.avatar.SetAlpha(float32(alpha))
	avatar.avatar.Draw(0, batch)
}

func (avatar *AvatarCircle) GetSprite() *sprite.Sprite {
	return avatar.avatar
}

func (avatar *AvatarCircle) GetWidth() float64 {
	return float64(avatar.avatar.Texture.Width * avatar.avatar.GetScale().X32())
}

func (avatar *AvatarCircle) loadAvatar(pixmap *texture.Pixmap) {
	avatarImage := pixmap.RGBA()
	padding := 25
	dst := image.NewRGBA(image.Rect(0, 0, pixmap.Width+padding, pixmap.Height+padding))
	center := dst.Bounds().Max.Div(2)
	circleRadius := pixmap.Width / 2
	draw.DrawMask(dst, dst.Bounds(), &image.Uniform{avatar.teamColor}, image.ZP, &circle{center, circleRadius + 10}, image.ZP, draw.Over)
	draw.DrawMask(dst, dst.Bounds(), avatarImage, image.Pt(-(padding/2), -(padding/2)), &circle{center, circleRadius}, image.ZP, draw.Over)

	tex := texture.LoadTextureSingle(dst, 4)
	region := tex.GetRegion()
	avatar.avatar = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(0, 0), vector.Centre)
	avatar.avatar.SetScale(float64(100 / region.Height))
}

func (avatar *AvatarCircle) LoadAvatarID(id int) {
	url := "https://a.ppy.sh/" + strconv.Itoa(id)

	log.Println("Trying to fetch avatar from:", url)

	request, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Println("Can't create request")
		return
	}

	client := new(http.Client)
	response, err := client.Do(request)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to create request to: \"%s\": %s", url, err))
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Println("a.ppy.sh responded with:", response.StatusCode)

		if response.StatusCode == 404 {
			log.Println("Avatar for user", id, "not found!")
		}

		return
	}

	pixmap, err := texture.NewPixmapReader(response.Body, response.ContentLength)
	if err != nil {
		log.Println("Can't load avatar! Error:", err)
		return
	}

	avatar.loadAvatar(pixmap)

	pixmap.Dispose()
}

func (avatar *AvatarCircle) LoadDefaultAvatar() {
	pixmap, err := assets.GetPixmap("assets/textures/dansercoin256.png")
	if err != nil {
		log.Println("Can't load avatar! Error:", err)
		return
	}

	avatar.loadAvatar(pixmap)

	pixmap.Dispose()
}

func (avatar *AvatarCircle) LoadAvatarUser(user string) {
	key, err := utils.GetApiKey()
	if err != nil {
		log.Println(fmt.Sprintf("Please put your osu!api v1 key into '%s' file", filepath.Join(env.ConfigDir(), "api.txt")))
	} else {
		client := osuapi.NewClient(key)
		err := client.Test()

		if err != nil {
			log.Println("Can't connect to osu!api:", err)
		} else {
			sUser, err := client.GetUser(osuapi.GetUserOpts{Username: user})
			if err != nil {
				log.Println("Can't find user:", user)
				log.Println(err)
			} else {
				avatar.LoadAvatarID(sUser.UserID)
			}
		}
	}
}
