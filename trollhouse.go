package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 800
const windowHeight = 600

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Configure the vertex and fragment shaders
	program, err := LoadShaderProgram("./shaders/test.vs", "./shaders/test.fs")
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Load the texture
	texture, err := LoadTexture("./resources/textures/square.png")
	if err != nil {
		log.Fatalln(err)
	}

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices)*4, gl.Ptr(cubeVertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 7*4, gl.PtrOffset(0))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 7*4, gl.PtrOffset(3*4))

	skinAttrib := uint32(gl.GetAttribLocation(program, gl.Str("skinAttr\x00")))
	gl.EnableVertexAttribArray(skinAttrib)
	gl.VertexAttribPointer(skinAttrib, 2, gl.FLOAT, false, 7*4, gl.PtrOffset(5*4))

	// Create animation tree
	node := NewAnimationNode([3]float32{0.0, -1.0, 0.0})
	node.addChild([3]float32{0.0, 1.0, 0.0})

	tree := AnimationTree{make([]*AnimationNode, 0), make([]SkinVertex, 0)}
	tree.addNodes(&node)

	// for i := 0; i < 6; i++ {
	// 	sv := SkinVertex{i, make(map[int]float32)}
	// 	sv.Weights[0] = 1.0
	// 	tree.Skin = append(tree.Skin, sv)
	// }

	// for i := 6; i < 12; i++ {
	// 	sv := SkinVertex{i, make(map[int]float32)}
	// 	sv.Weights[1] = 1.0
	// 	tree.Skin = append(tree.Skin, sv)
	// }

	anim1 := LoadAnimation("./resources/animations/twist.saf")
	anim2 := LoadAnimation("./resources/animations/bounce.saf")

	animationUniform := gl.GetUniformLocation(program, gl.Str("anim\x00"))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)

	// angle := 0.0
	previousTime := glfw.GetTime()

	anim1.begin(previousTime)
	anim2.begin(previousTime)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		// elapsed := time - previousTime
		previousTime = time

		// angle += elapsed
		// model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		// node.addTranslation([3]float32{0.0, 0.25 * float32(elapsed), 0.0})

		// gl.PolygonMode(GL_FRONT_AND_BACK, GL_LINE)

		// Render
		gl.UseProgram(program)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		tree.resetTree()
		tAnim := anim2.animate(anim1.animate(tree, time), time).getAnimation()
		fmt.Println(tAnim)
		gl.Uniform4fv(animationUniform,
			2,
			&(tAnim[0]))

		gl.BindVertexArray(vao)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.DrawArrays(gl.TRIANGLES, 0, 10*2*3)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func loadShaderSource(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	text := string(content) + "\x00"
	return text, nil
}

func LoadShaderProgram(vertexShaderFilename, fragmentShaderFilename string) (uint32, error) {
	vertexShaderSource, err := loadShaderSource(vertexShaderFilename)
	if err != nil {
		return 0, fmt.Errorf("shader %q not found on disk: %v", vertexShaderFilename, err)
	}

	fragmentShaderSource, err := loadShaderSource(fragmentShaderFilename)
	if err != nil {
		return 0, fmt.Errorf("shader %q not found on disk: %v", fragmentShaderFilename, err)
	}

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func LoadTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

var cubeVertices = []float32{
	//  X, Y, Z, U, V
	// Bottom
	-1.0, -1.0, -1.0, 0.0, 0.0, 0.0, -1.0,
	1.0, -1.0, -1.0, 1.0, 0.0, 0.0, -1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0, 0.0, -1.0,
	1.0, -1.0, -1.0, 1.0, 0.0, 0.0, -1.0,
	1.0, -1.0, 1.0, 1.0, 1.0, 0.0, -1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0, 0.0, -1.0,

	// Top
	-1.0, 1.0, -1.0, 0.0, 0.0, 1.0, -1.0,
	-1.0, 1.0, 1.0, 0.0, 1.0, 1.0, -1.0,
	1.0, 1.0, -1.0, 1.0, 0.0, 1.0, -1.0,
	1.0, 1.0, -1.0, 1.0, 0.0, 1.0, -1.0,
	-1.0, 1.0, 1.0, 0.0, 1.0, 1.0, -1.0,
	1.0, 1.0, 1.0, 1.0, 1.0, 1.0, -1.0,

	// Upper Front
	-1.0, 0.0, 1.0, 1.0, 0.0, 0.0, 1.0,
	1.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0, 1.0, -1.0,
	1.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 0.0, 1.0, 1.0, -1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0, 1.0, -1.0,

	// Lower Front
	-1.0, -1.0, 1.0, 1.0, 0.0, 0.0, -1.0,
	1.0, -1.0, 1.0, 0.0, 0.0, 0.0, -1.0,
	-1.0, 0.0, 1.0, 1.0, 1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 0.0, 0.0, 0.0, -1.0,
	1.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0,
	-1.0, 0.0, 1.0, 1.0, 1.0, 1.0, 0.0,

	// Upper Back
	-1.0, 0.0, -1.0, 0.0, 0.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 0.0, 1.0, 1.0, -1.0,
	1.0, 0.0, -1.0, 1.0, 0.0, 0.0, 1.0,
	1.0, 0.0, -1.0, 1.0, 0.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 0.0, 1.0, 1.0, -1.0,
	1.0, 1.0, -1.0, 1.0, 1.0, 1.0, -1.0,

	// Lower Back
	-1.0, -1.0, -1.0, 0.0, 0.0, 0.0, -1.0,
	-1.0, 0.0, -1.0, 0.0, 1.0, 1.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0, 0.0, -1.0,
	1.0, -1.0, -1.0, 1.0, 0.0, 0.0, -1.0,
	-1.0, 0.0, -1.0, 0.0, 1.0, 1.0, 0.0,
	1.0, 0.0, -1.0, 1.0, 1.0, 1.0, 0.0,

	// Upper Left
	-1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0, 1.0, -1.0,
	-1.0, 0.0, -1.0, 0.0, 0.0, 0.0, 1.0,
	-1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0, 1.0, -1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0, 1.0, -1.0,

	// Lower Left
	-1.0, -1.0, 1.0, 0.0, 1.0, 0.0, -1.0,
	-1.0, 0.0, -1.0, 1.0, 0.0, 1.0, 0.0,
	-1.0, -1.0, -1.0, 0.0, 0.0, 0.0, -1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0, 0.0, -1.0,
	-1.0, 0.0, 1.0, 1.0, 1.0, 1.0, 0.0,
	-1.0, 0.0, -1.0, 1.0, 0.0, 1.0, 0.0,

	// Upper Right
	1.0, 0.0, 1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 0.0, -1.0, 1.0, 0.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 0.0, 0.0, 1.0, -1.0,
	1.0, 0.0, 1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 0.0, 0.0, 1.0, -1.0,
	1.0, 1.0, 1.0, 0.0, 1.0, 1.0, -1.0,

	// Lower Right
	1.0, -1.0, 1.0, 1.0, 1.0, 0.0, -1.0,
	1.0, -1.0, -1.0, 1.0, 0.0, 0.0, -1.0,
	1.0, 0.0, -1.0, 0.0, 0.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0, 0.0, -1.0,
	1.0, 0.0, -1.0, 0.0, 0.0, 1.0, 0.0,
	1.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0,
}

type AnimationNode struct {
	Pos         [3]float32
	Translation [3]float32
	RotationY   float32
	Children    []*AnimationNode
}

func NewAnimationNode(pos [3]float32) AnimationNode {
	return AnimationNode{
		pos,
		[3]float32{0.0, 0.0, 0.0},
		0.0,
		make([]*AnimationNode, 0)}
}

func (p *AnimationNode) addChild(pos [3]float32) {
	n := NewAnimationNode(pos)
	(*p).Children = append((*p).Children, &n)
}

func (p *AnimationNode) translate(pos [3]float32) {
	(*p).Translation[0] += pos[0]
	(*p).Translation[1] += pos[1]
	(*p).Translation[2] += pos[2]

	for _, child := range (*p).Children {
		child.translate(pos)
	}
}

func (p *AnimationNode) resetTranslation() {
	(*p).Translation = [3]float32{0.0, 0.0, 0.0}
	for _, child := range (*p).Children {
		child.resetTranslation()
	}
}

func (p *AnimationNode) rotateY(rot float32) {
	(*p).RotationY += rot
	for _, child := range (*p).Children {
		child.rotateY(rot)
	}
}

func (p *AnimationNode) resetRotationY() {
	(*p).RotationY = 0.0
	for _, child := range (*p).Children {
		child.resetRotationY()
	}
}

type SkinVertex struct {
	VertexIdx int
	Weights   map[int]float32
}

type AnimationTree struct {
	Nodes []*AnimationNode
	Skin  []SkinVertex
}

func (t *AnimationTree) addNodes(node *AnimationNode) {
	(*t).Nodes = append((*t).Nodes, node)
	for _, child := range node.Children {
		(*t).addNodes(child)
	}
}

func (t AnimationTree) getAnimation() []float32 {
	ret := make([]float32, len(t.Nodes)*4)
	for i, node := range t.Nodes {
		ret[i*4] = node.Translation[0]
		ret[i*4+1] = node.Translation[1]
		ret[i*4+2] = node.Translation[2]
		ret[i*4+3] = node.RotationY
	}
	return ret
}

func (t *AnimationTree) resetTree() {
	for _, n := range (*t).Nodes {
		n.resetTranslation()
		n.resetRotationY()
	}
}

type NodeAnimationTranslation struct {
	NodeIdx     int
	Translation [3]float32
	RotationY   float32
}

type AnimationTimeStamp struct {
	TimePoint    int
	Translations []NodeAnimationTranslation
}

type Animation struct {
	StartTime         float64
	TimeStampDuration float32
	TimeStamps        []AnimationTimeStamp
}

func LoadAnimation(filename string) Animation {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var anim Animation
	anim.StartTime = 0.0
	anim.TimeStampDuration = 1.0
	anim.TimeStamps = make([]AnimationTimeStamp, 0)
	var timestamp AnimationTimeStamp
	first := true
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if words[0] == "ts" {
			if first {
				first = false
			} else {
				anim.TimeStamps = append(anim.TimeStamps, timestamp)
			}
			timestamp = AnimationTimeStamp{}

			ts, _ := strconv.Atoi(words[1])
			timestamp.TimePoint = ts
			timestamp.Translations = make([]NodeAnimationTranslation, 0)
		} else {
			var translation NodeAnimationTranslation
			translation.NodeIdx, _ = strconv.Atoi(words[0])

			var aux float64

			aux, _ = strconv.ParseFloat(words[1], 32)
			translation.Translation[0] = float32(aux)

			aux, _ = strconv.ParseFloat(words[2], 32)
			translation.Translation[1] = float32(aux)

			aux, _ = strconv.ParseFloat(words[3], 32)
			translation.Translation[2] = float32(aux)

			aux, _ = strconv.ParseFloat(words[4], 32)
			translation.RotationY = float32(aux)

			timestamp.Translations = append(timestamp.Translations, translation)
		}
	}
	anim.TimeStamps = append(anim.TimeStamps, timestamp)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return anim
}

func (a *Animation) begin(startTime float64) {
	(*a).StartTime = startTime
}

func (a Animation) animate(t AnimationTree, currTime float64) AnimationTree {
	// bound animation time
	time := float32(currTime - a.StartTime)
	for finalTimePoint := float32(a.TimeStamps[len(a.TimeStamps)-1].TimePoint) * a.TimeStampDuration; time > finalTimePoint; {
		time -= finalTimePoint
	}

	// get current run context
	var pos int
	for i, ts := range a.TimeStamps {
		pos = i
		if float32(ts.TimePoint)*a.TimeStampDuration > float32(time) {
			break
		}
	}

	ts := a.TimeStamps[pos]

	var prevTimePoint float32
	if pos > 0 {
		prevTimePoint = float32(a.TimeStamps[pos-1].TimePoint) * a.TimeStampDuration
	}

	factor := (time - prevTimePoint) / (float32(ts.TimePoint)*a.TimeStampDuration - prevTimePoint)
	// fmt.Printf("currTimePoint = %f, prevTimePoint = %f, time = %f, factor = %f\n", float32(ts.TimePoint)*a.TimeStampDuration, prevTimePoint, time, factor)

	for i, trans := range ts.Translations {
		currTrans := [3]float32{0.0, 0.0, 0.0}
		var currRotationY float32
		if pos > 0 {
			currTrans = vec3Lerp(a.TimeStamps[pos-1].Translations[i].Translation, trans.Translation, factor)
			currRotationY = lerp(a.TimeStamps[pos-1].Translations[i].RotationY, trans.RotationY, factor)
		} else {
			currTrans = vec3Lerp(currTrans, trans.Translation, factor)
			currRotationY = lerp(currRotationY, trans.RotationY, factor)
		}

		t.Nodes[trans.NodeIdx].translate(currTrans)
		t.Nodes[trans.NodeIdx].rotateY(currRotationY)
	}

	return t
}

// factor should be between 0 and 1
func lerp(a, b, factor float32) float32 {
	return a*(float32(1)-factor) + b*factor
}

func vec3Lerp(a, b [3]float32, factor float32) [3]float32 {
	return [3]float32{lerp(a[0], b[0], factor),
		lerp(a[1], b[1], factor),
		lerp(a[2], b[2], factor)}
}
