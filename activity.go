package jwtActivity

import (
	"errors"
	"fmt"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"github.com/mashling/mashling/registry"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

const (
	ivServiceName = "serviceName"
	ivToken 	= "token"
	ivKey        	= "key"
	ivSigningMethod = "signingMethod"
	ivIssuer      	= "iss"
	ivSubject       = "sub"
	ivAudience      = "aud"

	ovValid   	= "valid"
	ovToken 	= "token"
	ovValidationMsg = "validationMessage"
	ovError    	= "error"
	ovErrorMsg 	= "errorMessage"
)


type Factory struct {
}

func init() {
	registry.Register("jwtActivity", &Factory{})
}

type jwtActivity struct {
	metadata *activity.Metadata
}

// creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &jwtActivity{
		metadata: metadata,
	}
}

// Metadata return the metadata for the activity
func (f *jwtActivity) Metadata() *activity.Metadata {
	return f.metadata
}

// Eval executes the activity
func (f *jwtActivity) Eval(context activity.Context) (done bool, err error) {
	value := context.GetInput(ivServiceName)
	if value == nil {
		return false, errors.New("serviceName should not be nil")
	}
	serviceName, ok := value.(string)
	if !ok {
		return false, errors.New("serviceName should be a string")
	}

	settings := map[string]interface{}{
		ivToken:       context.GetInput(ivToken),
		ivKey: context.GetInput(ivKey),
		ivSigningMethod:     context.GetInput(ivSigningMethod),
		ivIssuer:        context.GetInput(ivIssuer),
		ivSubject:       context.GetInput(ivSubject),
		ivAudience:    context.GetInput(ivAudience),
	}
	factory := Factory{}
	service, err := factory.Make(serviceName, settings)
	if err != nil {
		return false, err
	}
	err = service.Execute()
	if err != nil {
		return false, err
	}
	context.SetOutput(ovValid, service.(*JWT).Response.Valid)
	context.SetOutput(ovToken, service.(*JWT).Response.Token)
	context.SetOutput(ovValidationMsg, service.(*JWT).Response.ValidationMessage)
	context.SetOutput(ovError, service.(*JWT).Response.Error)
	context.SetOutput(ovErrorMsg, service.(*JWT).Response.ErrorMessage)
	return true, nil
}

// JWT is a JWT validation service.
type JWT struct {
	Request  JWTRequest  `json:"request"`
	Response JWTResponse `json:"response"`
}

// JWTRequest is an JWT validation request.
type JWTRequest struct {
	Token         string `json:"token"`
	Key           string `json:"key"`
	SigningMethod string `json:"signingMethod"`
	Issuer        string `json:"iss"`
	Subject       string `json:"sub"`
	Audience      string `json:"aud"`
}

// JWTResponse is a parsed JWT response.
type JWTResponse struct {
	Valid             bool        `json:"valid"`
	Token             ParsedToken `json:"token"`
	ValidationMessage string      `json:"validationMessage"`
	Error             bool        `json:"error"`
	ErrorMessage      string      `json:"errorMessage"`
}

// ParsedToken is a parsed JWT token.
type ParsedToken struct {
	Claims        jwt.MapClaims          `json:"claims"`
	Signature     string                 `json:"signature"`
	SigningMethod string                 `json:"signingMethod"`
	Header        map[string]interface{} `json:"header"`
}

// InitializeHTTP initializes an HTTP service with provided settings.
func (f *Factory) Make(name string, settings map[string]interface{}) (registry.Service, error) {
	jwtService := &JWT{}
	request := JWTRequest{}
	jwtService.Request = request
	err := jwtService.setRequestValues(settings)
	if err != nil {
		fmt.Println("error in updaterequest")
	}
	return jwtService, err
}

// Execute invokes this JWT service.
func (j *JWT) Execute() error {
	j.Response = JWTResponse{}
	fmt.Println("Before passing to execute")
	fmt.Println("Signing Method", j.Request.SigningMethod)
	fmt.Println("Token being passed:", j.Request.Token)

	token, err := jwt.Parse(j.Request.Token, func(token *jwt.Token) (interface{}, error) {
		// Make sure signing alg matches what we expect
		switch strings.ToLower(j.Request.SigningMethod) {
		case "hmac":
			fmt.Println("in hmac")
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "ecdsa":
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "rsa":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "rsapss":
			if _, ok := token.Method.(*jwt.SigningMethodRSAPSS); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		case "":
		// Just continue
		default:
			return nil, fmt.Errorf("Unknown signing method expected: %v", j.Request.SigningMethod)
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if j.Request.Issuer != "" && !claims.VerifyIssuer(j.Request.Issuer, true) {
				fmt.Println("In Issuer")
				return nil, jwt.NewValidationError("iss claims do not match", jwt.ValidationErrorIssuer)
			}
			if j.Request.Audience != "" && !claims.VerifyAudience(j.Request.Audience, true) {
				fmt.Println("In Audience")
				return nil, jwt.NewValidationError("aud claims do not match", jwt.ValidationErrorAudience)
			}
			subClaim, sok := claims["sub"].(string)
			if j.Request.Subject != "" && (!sok || strings.Compare(j.Request.Subject, subClaim) != 0) {
				fmt.Println("In subject")
				return nil, jwt.NewValidationError("sub claims do not match", jwt.ValidationErrorClaimsInvalid)
			}
		} else {
			fmt.Println("in claims error")
			return nil, jwt.NewValidationError("unable to parse claims", jwt.ValidationErrorClaimsInvalid)
		}

		return []byte(j.Request.Key), nil
	})
	if token != nil && token.Valid {
		fmt.Println("valid token")
		j.Response.Valid = true
		j.Response.Token = ParsedToken{Signature: token.Signature, SigningMethod: token.Method.Alg(), Header: token.Header}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			j.Response.Token.Claims = claims
		}
		return err
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		fmt.Println("invalid token")
		j.Response.Valid = false
		j.Response.ValidationMessage = ve.Error()
	} else {
		fmt.Println("errror in token")
		j.Response.Valid = false
		j.Response.Error = true
		j.Response.ValidationMessage = err.Error()
		j.Response.ErrorMessage = err.Error()
	}
	return nil
}

// UpdateRequest updates a JWT validation service with new provided settings.
func (j *JWT) UpdateRequest(values map[string]interface{}) (err error) {
	return j.setRequestValues(values)
}

func (j *JWT) setRequestValues(settings map[string]interface{}) error {
	for k, v := range settings {
		if v == nil {
			continue
		}
		switch k {
		case "token":
			token, ok := v.(string)
			if !ok {
				return errors.New("invalid type for token")
			}
			// Try to scrub any extra noise from the token string
			fmt.Println("original token", token)
			tokenSplit := strings.Fields(token)
			fmt.Println("tokenSplit", tokenSplit)
			tokenSplit = strings.Split(tokenSplit[len(tokenSplit)-1],",")
			token = tokenSplit[0]
			token = token[:len(token)-1]
			fmt.Println("Assigned token", token)
			j.Request.Token = token
		case "key":
			key, ok := v.(string)
			if !ok {
				return errors.New("invalid type for key")
			}
			j.Request.Key = key
		case "signingMethod":
			signingMethod, ok := v.(string)
			if !ok {
				return errors.New("invalid type for signingMethod")
			}
			j.Request.SigningMethod = signingMethod
		case "issuer":
			issuer, ok := v.(string)
			if !ok {
				return errors.New("invalid type for issuer")
			}
			j.Request.Issuer = issuer
		case "subject":
			subject, ok := v.(string)
			if !ok {
				return errors.New("invalid type for subject")
			}
			j.Request.Subject = subject
		case "audience":
			audience, ok := v.(string)
			if !ok {
				return errors.New("invalid type for audience")
			}
			j.Request.Audience = audience
		default:
		// ignore and move on.
		}
	}
	return nil
}