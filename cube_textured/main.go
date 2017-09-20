// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Renders a textured spinning cube using GLFW 3 and OpenGL 4.1 core forward-compatible profile.
package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 1024
const windowHeight = 768

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube Texture", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err = gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Ensure we can capture the escape key being pressed below
	window.SetInputMode(glfw.StickyKeysMode, glfw.True)

	gl.ClearColor(0.0, 0.0, 0.4, 0)

	// Enable depth test
	gl.Enable(gl.DEPTH_TEST)
	// Accept fragment if it closer to the camera than the former one
	gl.DepthFunc(gl.LESS)

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))

	camera := mgl32.LookAt(4, 3, 3, 0, 0, 0, 0, 1, 0)
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))

	var cubeVerticies = []float32{
		-1.0, -1.0, -1.0, 0.000059, 0.000004,
		-1.0, -1.0, 1.0, 0.000103, 0.336048,
		-1.0, 1.0, 1.0, 0.335973, 0.335903,
		1.0, 1.0, -1.0, 1.000023, 0.000013,
		-1.0, -1.0, -1.0, 0.667979, 0.335851,
		-1.0, 1.0, -1.0, 0.999958, 0.336064,
		1.0, -1.0, 1.0, 0.667979, 0.335851,
		-1.0, -1.0, -1.0, 0.336024, 0.671877,
		1.0, -1.0, -1.0, 0.667969, 0.671889,
		1.0, 1.0, -1.0, 1.000023, 0.000013,
		1.0, -1.0, -1.0, 0.668104, 0.000013,
		-1.0, -1.0, -1.0, 0.667979, 0.335851,
		-1.0, -1.0, -1.0, 0.000059, 0.000004,
		-1.0, 1.0, 1.0, 0.335973, 0.335903,
		-1.0, 1.0, -1.0, 0.336098, 0.000071,
		1.0, -1.0, 1.0, 0.667979, 0.335851,
		-1.0, -1.0, 1.0, 0.335973, 0.335903,
		-1.0, -1.0, -1.0, 0.336024, 0.671877,
		-1.0, 1.0, 1.0, 1.000004, 0.671847,
		-1.0, -1.0, 1.0, 0.999958, 0.336064,
		1.0, -1.0, 1.0, 0.667979, 0.335851,
		1.0, 1.0, 1.0, 0.668104, 0.000013,
		1.0, -1.0, -1.0, 0.335973, 0.335903,
		1.0, 1.0, -1.0, 0.667979, 0.335851,
		1.0, -1.0, -1.0, 0.335973, 0.335903,
		1.0, 1.0, 1.0, 0.668104, 0.000013,
		1.0, -1.0, 1.0, 0.336098, 0.000071,
		1.0, 1.0, 1.0, 0.000103, 0.336048,
		1.0, 1.0, -1.0, 0.000004, 0.671870,
		-1.0, 1.0, -1.0, 0.336024, 0.671877,
		1.0, 1.0, 1.0, 0.000103, 0.336048,
		-1.0, 1.0, -1.0, 0.336024, 0.671877,
		-1.0, 1.0, 1.0, 0.335973, 0.335903,
		1.0, 1.0, 1.0, 0.667969, 0.671889,
		-1.0, 1.0, 1.0, 1.000004, 0.671847,
		1.0, -1.0, 1.0, 0.667979, 0.335851,
	}

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVerticies)*4, gl.Ptr(cubeVerticies), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))

	texture, err := newTexture("d6.png")
	if err != nil {
		log.Fatalln(err)
	}

	for !window.ShouldClose() && window.GetKey(glfw.KeyEscape) != glfw.Press {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render
		gl.UseProgram(program)

		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])
		gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}

	gl.DeleteBuffers(1, &vbo)
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteProgram(program)
}

var vertexShader = `
#version 330
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
in vec3 vert;
in vec2 vertTexCoord;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330
uniform sampler2D tex;
in vec2 fragTexCoord;
out vec4 outputColor;
void main() {
    outputColor = texture(tex, fragTexCoord);
}
` + "\x00"

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {

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

func newTexture(file string) (uint32, error) {
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
