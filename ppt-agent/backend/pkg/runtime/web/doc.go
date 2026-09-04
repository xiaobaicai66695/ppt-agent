// Package web is the compatibility façade and composition root for the Web
// layer. HTTP route registration lives in web/router, application operations
// live in web/service, and transport contracts live in web/model. Keeping the
// Server entry point here avoids breaking existing callers while the internal
// dependency direction remains router -> service/model.
package web
