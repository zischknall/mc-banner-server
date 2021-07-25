package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"strings"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/image/draw"
)

var height = 64
var width = height * 5
var lines = 1.0

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/minecraft/:server", minecraftHandler)

	err := app.Listen(":3000")
	if err != nil {
		return
	}
}

type status struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
	}
	Version struct {
		Name string
	}
	Favicon string
}

func minecraftHandler(c *fiber.Ctx) error {
	var s status
	err := pingServer(c.Params("server"), &s)
	if err != nil {
		return err
	}

	img, err := getFavicon(&s)
	if err != nil {
		return err
	}

	banner := generateBanner(img, &s)

	c.Set("Content-Type", "image/png")
	return banner.EncodePNG(c.Response().BodyWriter())
}

func generateBanner(favicon image.Image, s *status) *gg.Context {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0)
	dc.Clear()
	dc.SetRGB(1, 1, 1)
	dc.DrawImage(favicon, 0, 0)
	drawStringsToContext(dc, s.Description.ClearString())
	drawStringsToContext(dc, s.Version.Name)
	drawStringsToContext(dc, fmt.Sprintf("%d/%d", s.Players.Online, s.Players.Max))

	return dc
}

func getFavicon(s *status) (image.Image, error) {
	faviconString := strings.Split(s.Favicon, ",")[1]
	faviconData, err := base64.StdEncoding.DecodeString(faviconString)
	if err != nil {
		log.Printf("Unable to decode favicon base64: %v\n", err)
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewBuffer(faviconData))
	if err != nil {
		log.Printf("Decoding of favicon failed: %v\n", err)
		return nil, err
	}

	return scaleFavicon(img), nil
}

func scaleFavicon(old image.Image) image.Image {
	res := image.NewRGBA(image.Rect(0, 0, height, height))
	draw.BiLinear.Scale(res, res.Bounds(), old, old.Bounds(), draw.Over, nil)
	return res
}

func pingServer(address string, result *status) error {
	resp, _, err := bot.PingAndList(address)
	if err != nil {
		log.Printf("Ping and list failed: %v\n", err)
		return err
	}

	err = json.Unmarshal(resp, result)
	if err != nil {
		log.Printf("Unmarshal of ping and list response failed: %v\n", err)
		return err
	}

	return nil
}

func drawStringsToContext(dc *gg.Context, s string) {
	for _, n := range strings.Split(s, "\n") {
		n = strings.TrimSpace(n)
		dc.DrawStringAnchored(n, float64((width/2)+(height/2)), dc.FontHeight()*lines, 0.5, 0.5)
		lines += 1
	}
}
