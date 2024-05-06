package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/image/colornames"
	constant "issue_pr_board/utils/aj-captcha/const"
	"issue_pr_board/utils/aj-captcha/model/vo"
	"issue_pr_board/utils/aj-captcha/util"
	img "issue_pr_board/utils/aj-captcha/util/image"
	"log"
	"math"
)

type BlockPuzzleCaptchaService struct {
	point   vo.PointVO
	factory *CaptchaServiceFactory
}

func NewBlockPuzzleCaptchaService(factory *CaptchaServiceFactory) *BlockPuzzleCaptchaService {
	img.SetUp(factory.config.ResourcePath)
	return &BlockPuzzleCaptchaService{
		factory: factory,
	}
}

func (b *BlockPuzzleCaptchaService) Get() (map[string]interface{}, error) {
	backgroundImage := img.GetBackgroundImage()
	templateImage := img.GetTemplateImage()
	b.pictureTemplatesCut(backgroundImage, templateImage)

	originalImageBase64, err := backgroundImage.Base64()
	jigsawImageBase64, err := templateImage.Base64()

	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["originalImageBase64"] = originalImageBase64
	data["jigsawImageBase64"] = jigsawImageBase64
	data["secretKey"] = b.point.SecretKey
	data["token"] = util.GetUuid()

	codeKey := fmt.Sprintf(constant.CodeKeyPrefix, data["token"])
	jsonPoint, err := json.Marshal(b.point)
	if err != nil {
		log.Printf("point json Marshal err: %v", err)
		return nil, err
	}

	b.factory.GetCache().Set(codeKey, string(jsonPoint), b.factory.config.CacheExpireSec)

	return data, nil
}

func (b *BlockPuzzleCaptchaService) pictureTemplatesCut(backgroundImage *util.ImageUtil, templateImage *util.ImageUtil) {
	b.generateJigsawPoint(backgroundImage, templateImage)
	b.cutByTemplate(backgroundImage, templateImage, b.point.X, 0)
}

func (b *BlockPuzzleCaptchaService) cutByTemplate(backgroundImage *util.ImageUtil, templateImage *util.ImageUtil, x1, y1 int) {
	xLength := templateImage.Width
	yLength := templateImage.Height

	for x := 0; x < xLength; x++ {
		for y := 0; y < yLength; y++ {
			isOpacity := templateImage.IsOpacity(x, y)

			backgroundX := x + x1
			backgroundY := y + y1

			if !isOpacity {
				backgroundRgba := backgroundImage.RgbaImage.RGBAAt(backgroundX, backgroundY)
				templateImage.SetPixel(backgroundRgba, x, y)
				backgroundImage.VagueImage(backgroundX, backgroundY)
			}

			if x == (xLength-1) || y == (yLength-1) {
				continue
			}

			rightOpacity := templateImage.IsOpacity(x+1, y)
			downOpacity := templateImage.IsOpacity(x, y+1)

			if (isOpacity && !rightOpacity) || (!isOpacity && rightOpacity) || (isOpacity && !downOpacity) || (!isOpacity && downOpacity) {
				templateImage.RgbaImage.SetRGBA(x, y, colornames.White)
				backgroundImage.RgbaImage.SetRGBA(backgroundX, backgroundY, colornames.White)
			}
		}
	}
}

func (b *BlockPuzzleCaptchaService) generateJigsawPoint(backgroundImage *util.ImageUtil, templateImage *util.ImageUtil) {
	widthDifference := backgroundImage.Width - templateImage.Width
	heightDifference := backgroundImage.Height - templateImage.Height

	x, y := 0, 0

	if widthDifference <= 0 {
		x = 5
	} else {
		x = util.RandomInt(100, widthDifference-100)
	}
	if heightDifference <= 0 {
		y = 5
	} else {
		y = util.RandomInt(5, heightDifference)
	}
	point := vo.PointVO{X: x, Y: y}
	point.SetSecretKey(util.RandString(16))
	b.point = point
}

func (b *BlockPuzzleCaptchaService) Check(token string, pointJson string) error {
	cache := b.factory.GetCache()

	codeKey := fmt.Sprintf(constant.CodeKeyPrefix, token)

	cachePointInfo := cache.Get(codeKey)

	if cachePointInfo == "" {
		return errors.New("验证码已失效")
	}

	cachePoint := &vo.PointVO{}
	userPoint := &vo.PointVO{}
	err := json.Unmarshal([]byte(cachePointInfo), cachePoint)

	if err != nil {
		return err
	}

	userPointJson := util.AesDecrypt(pointJson, cachePoint.SecretKey)

	err = json.Unmarshal([]byte(userPointJson), userPoint)

	if err != nil {
		return err
	}

	if math.Abs(float64(cachePoint.X-userPoint.X)) <= float64(b.factory.config.BlockPuzzle.Offset) && cachePoint.Y == userPoint.Y {
		return nil
	}

	return errors.New("验证失败")
}

func (b *BlockPuzzleCaptchaService) Verification(token string, pointJson string) error {
	err := b.Check(token, pointJson)
	if err != nil {
		return err
	}
	codeKey := fmt.Sprintf(constant.CodeKeyPrefix, token)
	b.factory.GetCache().Delete(codeKey)
	return nil
}
