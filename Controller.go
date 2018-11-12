package pgo

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "reflect"

    "github.com/pinguo/pgo/Util"
)

// Controller the base class of web and cmd controller
type Controller struct {
    Object
    Status int    // response status
    Output []byte // response output
}

// GetBindInfo get action map as extra binding info
func (c *Controller) GetBindInfo(v interface{}) interface{} {
    if _, ok := v.(IController); !ok {
        panic("param require a controller")
    }

    rt := reflect.ValueOf(v).Type()
    num := rt.NumMethod()
    actions := make(map[string]int)

    for i := 0; i < num; i++ {
        name := rt.Method(i).Name
        if len(name) > ActionLength && name[:ActionLength] == ActionPrefix {
            actions[name[ActionLength:]] = i
        }
    }

    return actions
}

// BeforeAction before action hook
func (c *Controller) BeforeAction(action string) {
}

// AfterAction after action hook
func (c *Controller) AfterAction(action string) {
}

// FinishAction finish action hook
func (c *Controller) FinishAction(action string) {
    ctx := c.GetContext()

    ctx.Notice("[%d(ms)] [1(MB)] [%s] [%s] profile[%s] counting[%s]",
        ctx.GetElapseMs(),
        ctx.GetPath(),
        ctx.GetPushLogString(),
        ctx.GetProfileString(),
        ctx.GetCountingString(),
    )

    ctx.Profiler.Reset()

    // send output at the end of the controller's lifecycle,
    // give opportunity to modify output in AfterAction hook
    if c.Status != 0 || c.Output != nil {
        ctx.End(c.Status, c.Output)
    }
}

// HandlePanic process unhandled action panic
func (c *Controller) HandlePanic(v interface{}) {
    status := http.StatusInternalServerError
    switch e := v.(type) {
    case *Exception:
        status = e.GetStatus()
        c.OutputJson(EmptyObject, status, e.GetMessage())
    default:
        c.OutputJson(EmptyObject, status)
    }

    if !App.GetServer().IsErrorLogOff(status) {
        c.GetContext().Error("%s, trace[%s]", Util.ToString(v), Util.PanicTrace(TraceMaxDepth, false))
    }
}

// Redirect output redirect response
func (c *Controller) Redirect(location string, permanent bool) {
    if permanent {
        c.Status = http.StatusMovedPermanently
    } else {
        c.Status = http.StatusFound
    }

    c.GetContext().SetHeader("Location", location)
}

// OutputJson output json response
func (c *Controller) OutputJson(data interface{}, status int, msg ...string) {
    message := App.GetStatus().GetText(status, c.GetContext(), msg...)
    output, e := json.Marshal(map[string]interface{}{
        "status":  status,
        "message": message,
        "data":    data,
    })

    if e != nil {
        panic(fmt.Sprintf("failed to marshal json, %s", e))
    }

    c.Status = http.StatusOK
    c.Output = output

    c.GetContext().PushLog("status", status)
    c.GetContext().SetHeader("Content-Type", "application/json; charset=utf-8")
}

// OutputJsonp output jsonp response
func (c *Controller) OutputJsonp(callback string, data interface{}, status int, msg ...string) {
    message := App.GetStatus().GetText(status, c.GetContext(), msg...)
    output, e := json.Marshal(map[string]interface{}{
        "status":  status,
        "message": message,
        "data":    data,
    })

    if e != nil {
        panic(fmt.Sprintf("failed to marshal json, %s", e))
    }

    buf := &bytes.Buffer{}
    buf.WriteString(callback + "(")
    buf.Write(output)
    buf.WriteString(")")

    c.Status = http.StatusOK
    c.Output = buf.Bytes()

    c.GetContext().PushLog("status", status)
    c.GetContext().SetHeader("Content-Type", "text/javascript; charset=utf-8")
}

// OutputView output rendered view
func (c *Controller) OutputView(view string, data interface{}) {
    c.Status = http.StatusOK
    c.Output = App.GetView().Render(view, data)
    c.GetContext().PushLog("status", c.Status)
    c.GetContext().SetHeader("Content-Type", "text/html; charset=utf-8")
}
