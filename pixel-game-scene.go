package pixel-framework

import (
	"time"
	"math/rand"
	"math"
	"fmt"
	"image/color"
	"log/slog"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
)


const BOARD_SIZE		= 8
const TILE_WIDTH		= 132.0
const TILE_HEIGHT		= 66.0
const ALPHA_FADE_AMOUNT	= 2
const FADEIN_TIME		= 100
const ACHIEVEMENT_TIME	= 600

type GameState 		int
type MenuSelection 	int

const (
	GameState_Title = iota
	GameState_Settings
	GameState_About
	GameState_ReallyQuit
	GameState_Running
	GameState_GameOver
)

const (
	MenuSelection_NewGame = iota
	MenuSelection_Settings
	MenuSelection_About
	MenuSelection_Quit
)

var MenuSelectionName = map[MenuSelection]string{
	MenuSelection_NewGame:	"New Game",
	MenuSelection_Settings:	"Settings",
	MenuSelection_About:	"About",
	MenuSelection_Quit:		"Quit",
}

type CityGame_Tile struct {
	sprite				*pixel.Sprite
	details				*SpriteSheet_SubTexture
	highlightColor		color.RGBA
	highlightTile		bool
}

type CityGame_Elements struct {
	win							opengl.Window
	textAtlas					*text.Atlas
	debugAtlas					*text.Atlas

	// Polygon Drawing
	imdraw						*imdraw.IMDraw
	achievementDraw				*imdraw.IMDraw

	// Text Labels
	fpsText						*text.Text
	scoreText					*text.Text					
	debugText					*text.Text
	mousePosText				*text.Text
	achievementNameText			*text.Text
	achievementDescText			*text.Text

	// Tile Data
	spritesheet					*pixel.Picture
	textureAtlas				*SpriteSheet_TextureAtlas
	tileBatch					*pixel.Batch
	
	// Current Tile Data
	boardTiles					[BOARD_SIZE][BOARD_SIZE]CityGame_Tile
	currentTileGroup			int
	currentTile					int

	// Display Tracking
	camPos						pixel.Vec
	camSpeed					float64
	camZoom						float64
	camZoomSpeed				float64

	frames						int
	second						<-chan time.Time
	debugMode					bool
	
	// Game State
	score						int
	state						GameState			
	menuSelection				MenuSelection
} 

func(el *CityGame_Elements) _generateTextElement(pos pixel.Vec, face *text.Atlas) *text.Text {
	item := text.New(pos, face)
	item.Color = colornames.White
	return item
}

func(el *CityGame_Elements) _initTextElements() error {

	face, err := loadTTF("resources/fonts/Chibi.ttf", 40)
	if err != nil {
		panic(err)
	}

	el.textAtlas = text.NewAtlas(face, text.ASCII)

	debugFace, err := loadTTF("resources/fonts/Chibi.ttf", 15)
	if err != nil {
		panic(err)
	}

	el.debugAtlas = text.NewAtlas(debugFace, text.ASCII)


	el.fpsText = el._generateTextElement(pixel.V(-60, 250), el.debugAtlas)
	el.scoreText = el._generateTextElement(pixel.V(-60, 400), el.textAtlas)
	el.debugText = el._generateTextElement(pixel.V(-60, 350), el.debugAtlas)
	el.mousePosText = el._generateTextElement(pixel.V(-60, 300), el.debugAtlas)
	el.achievementNameText = el._generateTextElement(pixel.V(-90, -210), el.textAtlas)
	el.achievementDescText = el._generateTextElement(pixel.V(-90, -235), el.debugAtlas)

	return nil
}

func(el *CityGame_Elements) _initTileElements() error {

	ss, err := loadPicture("resources/images/cityTiles_sheet.png")
	if err != nil {
		return err
	}
	el.spritesheet = &ss
	ta, err := loadTextureAtlas("resources/images/cityTiles_sheet.xml")
	if err != nil {
		return err
	}
	el.textureAtlas = &ta
	el.tileBatch = pixel.NewBatch(&pixel.TrianglesData{}, *el.spritesheet)

	return nil
}

func(el *CityGame_Elements) _initDrawingElements() {
	el.imdraw = imdraw.New(nil)
	el.achievementDraw = imdraw.New(nil)

	el.imdraw.Color = pixel.RGB(0.0, 0.2, 0.0)
	for x := 0; x <= BOARD_SIZE; x++ {
		p1 := el.mapTilePos(x, 0).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		p2 := el.mapTilePos(x, BOARD_SIZE).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		el.imdraw.Push(p1)
		el.imdraw.Push(p2)
		el.imdraw.Line(3)
	} 

	for y := 0; y <= BOARD_SIZE; y++ {
		p1 := el.mapTilePos(0, y).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		p2 := el.mapTilePos(BOARD_SIZE, y).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		el.imdraw.Push(p1)
		el.imdraw.Push(p2)
		el.imdraw.Line(3)
	} 
}

func(el *CityGame_Elements) mapTilePos(x int, y int) pixel.Vec {
	return pixel.V(
		(float64(x) * TILE_WIDTH / 2.0) + (float64(y) * TILE_WIDTH / 2.0),
		(-1.0 * float64(x) * TILE_HEIGHT / 2.0) + (float64(y) * TILE_HEIGHT / 2.0))
}

func(el *CityGame_Elements) cycleNextTile() {
	el.currentTile++
	if el.currentTile >= len(el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures) {
		el.currentTile = 0
	}
}

func(el *CityGame_Elements) getNextTileGroup() {
	// Actual Game
	el.currentTileGroup = rand.Intn(len(el.textureAtlas.TileGroups))
	// Debugging
	// el.currentTileGroup++
	// if el.currentTileGroup >= len(el.textureAtlas.TileGroups) {
	// 	el.currentTileGroup = 0
	// }

	el.currentTile = 0
}

func(el *CityGame_Elements) cycleTileGroup() {
	el.currentTileGroup++
	if el.currentTileGroup >= len(el.textureAtlas.TileGroups) {
		el.currentTileGroup = 0
	}

	el.currentTile = 0
}

func(el *CityGame_Elements) getCurrentTile() *CityGame_Tile {
	// Get the current tile bounds and create a sprite
	texture := el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile]
	tileBounds := pixel.R(
		(*el.spritesheet).Bounds().Min.X + float64(texture.X), 
		(*el.spritesheet).Bounds().Max.Y - float64(texture.Y) - float64(texture.Height), 
		(*el.spritesheet).Bounds().Min.X + float64(texture.X + texture.Width), 
		(*el.spritesheet).Bounds().Max.Y - float64(texture.Y))
	tileSprite := pixel.NewSprite(*el.spritesheet, tileBounds)
	tile := &CityGame_Tile{
		sprite:			tileSprite,
		details:		&el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile],
	}

	return tile
}

func(el *CityGame_Elements) getCurrentTileGroupName() string {
	return el.textureAtlas.TileGroups[el.currentTileGroup].Name
}

func(el *CityGame_Elements) getCurrentTileJoinTL() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinTL
}

func(el *CityGame_Elements) getCurrentTileJoinTR() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinTR
}

func(el *CityGame_Elements) getCurrentTileJoinBR() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinBR
}

func(el *CityGame_Elements) getCurrentTileJoinBL() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinBL
}

func(el *CityGame_Elements) getTotalTileCount() int {
	ret := 0
	for x := 0; x < BOARD_SIZE; x++ {
		for y := 0; y < BOARD_SIZE; y++ {
			if el.boardTiles[x][y].sprite != nil {
				ret++
			}
		}
	}
	return ret
}

func (el *CityGame_Elements) checkAddCurrentTile (x int, y int, tile *CityGame_Tile) {

	if el.boardTiles[x][y].sprite == nil {
		el.boardTiles[x][y] = *tile

		// Score Checks!
		thisTile := &el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile]

		if (x > 0 && el.boardTiles[x - 1][y].sprite != nil && 
			thisTile.JoinTL.compatibleWith(el.boardTiles[x - 1][y].details.JoinBR)) ||
		   (x < (BOARD_SIZE - 1) && el.boardTiles[x + 1][y].sprite != nil && 
			thisTile.JoinBR.compatibleWith(el.boardTiles[x + 1][y].details.JoinTL)) ||
		   (y > 0 && el.boardTiles[x][y - 1].sprite != nil && 
			thisTile.JoinBL.compatibleWith(el.boardTiles[x][y - 1].details.JoinTR)) ||
		   (y < (BOARD_SIZE - 1) && el.boardTiles[x][y + 1].sprite != nil && 
			thisTile.JoinTR.compatibleWith(el.boardTiles[x][y + 1].details.JoinBL)) {
			el.score++
			el.boardTiles[x][y].highlightTile = true
			el.boardTiles[x][y].highlightColor = colornames.Yellow
			el.boardTiles[x][y].highlightColor.A = 0
		}

		el.checkOtherAchievements(x, y)
	
		if el.getTotalTileCount() == (BOARD_SIZE * BOARD_SIZE) {
			panic(fmt.Sprintf("GAME OVER: Score = %d", el.score))
		}

		el.getNextTileGroup()
	}
}

func(el *CityGame_Elements) checkOtherAchievements(x int, y int) {
	slog.Info("--------------Checking for other achievements--------------")

	el.checkCompletedAchievements(x, y)
	

	// Special Cases
	el.checkTunnelAchievements(x, y)



	slog.Info("--------------Done with Checks--------------")
	slog.Info("    ")
	
}

func(el *CityGame_Elements) checkTunnelAchievements(x int, y int) {
	// Is this a tunnel?
	if el.boardTiles[x][y].details.JoinTL == JoinType_TunnelHeight || 
		el.boardTiles[x][y].details.JoinTL == JoinType_TunnelDepth {
		el.checkTunnelAchievementsInDirection(x, y, -1, 0, el.boardTiles[x][y].details.JoinTL)
	} else if el.boardTiles[x][y].details.JoinTR == JoinType_TunnelHeight || 
		el.boardTiles[x][y].details.JoinTR == JoinType_TunnelDepth {
		el.checkTunnelAchievementsInDirection(x, y, 0, 1, el.boardTiles[x][y].details.JoinTR)
	} else if el.boardTiles[x][y].details.JoinBR == JoinType_TunnelHeight || 
		el.boardTiles[x][y].details.JoinBR == JoinType_TunnelDepth {
		el.checkTunnelAchievementsInDirection(x, y, 1, 0, el.boardTiles[x][y].details.JoinBR)
	} else if el.boardTiles[x][y].details.JoinBL == JoinType_TunnelHeight || 
		el.boardTiles[x][y].details.JoinBL == JoinType_TunnelDepth {
		el.checkTunnelAchievementsInDirection(x, y, 0, -1, el.boardTiles[x][y].details.JoinBL)
	}
}

func(el *CityGame_Elements) checkTunnelAchievementsInDirection(x int, y int, x_off int, y_off int, joinType JoinType) {
	x_search := x
	y_search := y
	len := 0
	for x_search >= 0 && x_search < BOARD_SIZE && y_search >= 0 && y_search < BOARD_SIZE {

		if el.boardTiles[x_search][y_search].details != nil {
			if (x_off == -1 && y_off == 0 && el.boardTiles[x_search][y_search].details.JoinBR == joinType) ||
				(x_off == 1 && y_off == 0 && el.boardTiles[x_search][y_search].details.JoinTL == joinType) ||
				(x_off == 0 && y_off == -1 && el.boardTiles[x_search][y_search].details.JoinTR == joinType) ||
				(x_off == 0 && y_off == 0 && el.boardTiles[x_search][y_search].details.JoinBL == joinType) {
				UnlockAchievement_TunnelSize(len)
			}
		}
		x_search += x_off
		y_search += y_off
		len++
	}
}

func(el *CityGame_Elements) checkCompletedAchievements(x int, y int) {

	// Check whether this thing makes a complete group
	//
	var completed [BOARD_SIZE][BOARD_SIZE]bool

	// Algorithm:
	// Check each edge - if it's a relevant join, check that the join is good.
	// If any not good, return false
	// Otherwise mark the cell as visited and then go to each join and check that cell

	if el.checkCell(x, y, &completed) {
		row_str := ""
		accum := 2
		totalTiles := 0
		for check_y := BOARD_SIZE - 1; check_y >= 0; check_y-- {
			row_str = ""
			for check_x := 0; check_x < BOARD_SIZE; check_x++ {
				if completed[check_x][check_y] {

					accum += 1
					totalTiles += 1
					el.score += accum
					el.boardTiles[check_x][check_y].highlightTile = true
					el.boardTiles[check_x][check_y].highlightColor = colornames.Purple
					el.boardTiles[check_x][check_y].highlightColor.A = 0

					row_str = row_str + "*"
				} else {
					row_str = row_str + " "
				}
			}
			slog.Info(row_str)
		}
		if totalTiles > 0 {
			switch el.textureAtlas.TileGroups[el.currentTileGroup].Type {
			case "road":
				UnlockAchievement_RoadSize(totalTiles)
			case "river":
				UnlockAchievement_RiverSize(totalTiles)
			case "park":
				UnlockAchievement_ParkSize(totalTiles)
			}
		}
	}
}

func(el *CityGame_Elements) checkCell(x int, y int, completed*[BOARD_SIZE][BOARD_SIZE] bool) bool {

	var cellsToCheck [4]bool
	if completed[x][y] {
		// We've already checked this one
		slog.Info("CHECKCELLS: Already complete", "x", x, "y", y)
		return true
	}
	completed[x][y] = el.checkAllJoins(x, y, &cellsToCheck)
	slog.Info("Checking: ", "x", x, "y", y, "completed", completed[x][y], "cellsToCheck", cellsToCheck)
	if completed[x][y] {
		// Check all surrounding
		if !cellsToCheck[0] || (x == 0 || el.checkCell(x - 1, y, completed)) {
			if !cellsToCheck[1] || (x == BOARD_SIZE - 1 || el.checkCell(x + 1, y, completed)) {
				if !cellsToCheck[2] || (y == 0 || el.checkCell(x, y - 1, completed)) {
					if !cellsToCheck[3] || (y == BOARD_SIZE - 1 || el.checkCell(x, y + 1, completed)) {
						return true
					}
				}
			}
		}
	}
	return false
}

func(el *CityGame_Elements) checkAllJoins(x int, y int, checkConnections*[4] bool) bool {
	if el.boardTiles[x][y].details != nil {
		// TL
		var tlComplete = false
		if (el.boardTiles[x][y].details.JoinTL == JoinType_Concrete || 
			x == 0) {
			tlComplete = true
		} else if 
			(x > 0 && el.boardTiles[x - 1][y].sprite != nil &&
			el.boardTiles[x][y].details.JoinTL.compatibleWith(el.boardTiles[x - 1][y].details.JoinBR)) {
			tlComplete = true
			slog.Info("****** SHOULD BE GOOOOOD")
			checkConnections[0] = true
		}
		slog.Info("    TL: ", "x", x, "y", y, "tlComplete", tlComplete)
			
		if tlComplete {
			//slog.Info("Left is complete!")

			// BR
			var brComplete = false
			if (el.boardTiles[x][y].details.JoinBR == JoinType_Concrete || (x == (BOARD_SIZE - 1))) {
				brComplete = true
			} else if 
				(x < (BOARD_SIZE - 1) && el.boardTiles[x + 1][y].sprite != nil &&
				el.boardTiles[x][y].details.JoinBR.compatibleWith(el.boardTiles[x + 1][y].details.JoinTL)) {
				brComplete = true
				slog.Info("****** SHOULD BE GOOOOOD2")
				checkConnections[1] = true
			}
			slog.Info("    BR: ", "x", x, "y", y, "brComplete", brComplete)
			
			if brComplete {
				//slog.Info("Right is complete!")
				
				// BL
				var blComplete = false
				if (el.boardTiles[x][y].details.JoinBL == JoinType_Concrete || y == 0) {
					blComplete = true
				} else if 
					(y > 0 && el.boardTiles[x][y - 1].sprite != nil &&
					el.boardTiles[x][y].details.JoinBL.compatibleWith(el.boardTiles[x][y - 1].details.JoinTR)) {
					blComplete = true
					slog.Info("****** SHOULD BE GOOOOOD3")
					checkConnections[2] = true
				}
				slog.Info("    BL: ", "x", x, "y", y, "blComplete", blComplete)
				
				if blComplete {
					//slog.Info("Bottom is complete!")

					// TR
					var trComplete = false
					if (el.boardTiles[x][y].details.JoinTR == JoinType_Concrete || ( y == (BOARD_SIZE - 1))) {
						trComplete = true
					} else if 
						(y < (BOARD_SIZE - 1) && el.boardTiles[x][y + 1].sprite != nil &&
						el.boardTiles[x][y].details.JoinTR.compatibleWith(el.boardTiles[x][y + 1].details.JoinBL)) {
						trComplete = true
						slog.Info("****** SHOULD BE GOOOOOD4")
						checkConnections[3] = true
					}
					slog.Info("    TR: ", "x", x, "y", y, "trComplete", trComplete)

					if trComplete {
						slog.Info("------> Connections? ", "checkConnections", checkConnections)
						//slog.Info("Top is complete!")

						return true;
					}
				}
			}
		}
	}
	return false
}

func(el *CityGame_Elements) scrollLeft(dt float64) {
	el.camPos.X -= el.camSpeed * dt
}

func(el *CityGame_Elements) scrollRight(dt float64) {
	el.camPos.X += el.camSpeed * dt
}

func(el *CityGame_Elements) scrollUp(dt float64) {
	el.camPos.Y += el.camSpeed * dt
}

func(el *CityGame_Elements) scrollDown(dt float64) {
	el.camPos.Y -= el.camSpeed * dt
}

func(el *CityGame_Elements) zoom(zoomLevel float64) {
	el.camZoom *= math.Pow(el.camZoomSpeed, zoomLevel)
}

func(el *CityGame_Elements) resetGame() {
	el.score = 0
	for x := 0; x < BOARD_SIZE; x++ {
		for y := 0; y < BOARD_SIZE; y++ {
			el.boardTiles[x][y].details = nil
			el.boardTiles[x][y].sprite = nil
		}
	}
	el.getNextTileGroup()
}


func(el *CityGame_Elements) doFades() {
	for x := 0; x < BOARD_SIZE; x++ {
		for y := 0; y < BOARD_SIZE; y++ {
			if el.boardTiles[x][y].highlightTile {
				if el.boardTiles[x][y].highlightColor.A < (0xFF - ALPHA_FADE_AMOUNT) {
					el.boardTiles[x][y].highlightColor.A += ALPHA_FADE_AMOUNT
				} else {
					el.boardTiles[x][y].highlightTile = false
				}
			}
		}
	}
}

func InitGameElements() (*CityGame_Elements, error) {
	e := &CityGame_Elements{
		state:				GameState_Title,
		menuSelection:		MenuSelection_NewGame,
		currentTileGroup: 	0,
		currentTile: 		0,
		camPos: 			pixel.V(480.0, 120.0),
		camSpeed: 			500.0,
		camZoom:			1.0,
		camZoomSpeed:		1.2,
		frames:				0,
		second:				time.Tick(time.Second),
		debugMode:			false,

		score:				0,
	}

	err := e._initTextElements()
	if err != nil {
		return nil, err
	}

	err = e._initTileElements()
	if err != nil {
		return nil, err
	}

	e._initDrawingElements()

	e.getNextTileGroup()
	
	return e, nil
}
