// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/meal/ids": {
            "get": {
                "description": "MealID 배열인 data를 가진 구조체를 리턴받는다.",
                "produces": [
                    "application/json"
                ],
                "summary": "학식 게시판에서 학식 목록을 가져온다.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.GetMealIdsResult"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/main.ErrorMessage"
                        }
                    },
                    "502": {
                        "description": "Bad Gateway",
                        "schema": {
                            "$ref": "#/definitions/main.ErrorMessage"
                        }
                    }
                }
            }
        },
        "/schedules/{year}/{month}": {
            "get": {
                "description": "ScheduleItem 배열인 schedules를 가진 구조체를 리턴받는다.",
                "produces": [
                    "application/json"
                ],
                "summary": "월간 학사 일정 조회",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "년도",
                        "name": "year",
                        "in": "path"
                    },
                    {
                        "type": "integer",
                        "description": "월",
                        "name": "month",
                        "in": "path"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.GetSchedulesResult"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/main.ErrorMessage"
                        }
                    },
                    "502": {
                        "description": "Bad Gateway",
                        "schema": {
                            "$ref": "#/definitions/main.ErrorMessage"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "main.ErrorMessage": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "main.GetMealIdsResult": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/main.MealID"
                    }
                }
            }
        },
        "main.GetSchedulesResult": {
            "type": "object",
            "properties": {
                "schedules": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/main.ScheduleItem"
                    }
                }
            }
        },
        "main.MealID": {
            "type": "object",
            "properties": {
                "date": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "main.ScheduleItem": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "period": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
