// Copyright 2016 aletheia7. All rights reserved. Use of this source code is
// governed by a BSD-2-Clause license that can be found in the LICENSE file.
// +build linux

// Package sd provides methods to write to the systemd-journal.
package sd

/*
New_journal() and New_journal_m() create a Journal struct.
Journal.Emerg(), Journal.Alert(), Journal.Crit(), Journal.Err(),
Journal.Warning(), Journal.Notice(), Journal.Info(), Journal.Debug() write
to the systemd journal.

Each method contains a *_m (map variation) method that allows sending your
own fields. The map suppports string and []byte (binary). Each method also
contains a _m_f (map & format variation) method that supports
http://godoc.org/fmt#Printf style arguments.

Each method contains a *_a (array variation) method that allows
sending your own fields as []string{"FIELD1=v1", "FIELD2=v2"}.
Each method
also contains a _a_f (array & format variation) method that supports
http://godoc.org/fmt#Printf style arguments.

Each of the methods will add journal fields GO_LINE, and GO_FUNC fields to
the journal to indicate where the methods were called. The *_m_f methods
can take nil map in order to only use the format functionality.
*/

/*
#cgo pkg-config: libsystemd
#include <stdlib.h>
#include <systemd/sd-journal.h>
#include <unistd.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aletheia7/sd/v6/ansi"
	"io"
	"log/syslog"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

type Priority string

// These are log/syslog.Priority values.
var (
	Log_emerg   = Priority(strconv.Itoa(int(syslog.LOG_EMERG)))
	Log_alert   = Priority(strconv.Itoa(int(syslog.LOG_ALERT)))
	Log_crit    = Priority(strconv.Itoa(int(syslog.LOG_CRIT)))
	Log_err     = Priority(strconv.Itoa(int(syslog.LOG_ERR)))
	Log_warning = Priority(strconv.Itoa(int(syslog.LOG_WARNING)))
	Log_notice  = Priority(strconv.Itoa(int(syslog.LOG_NOTICE)))
	Log_info    = Priority(strconv.Itoa(int(syslog.LOG_INFO)))
	Log_debug   = Priority(strconv.Itoa(int(syslog.LOG_DEBUG)))
)

const (
	sd_go_func  = "GO_FUNC"
	sd_go_file  = "GO_FILE"
	sd_priority = "PRIORITY"
	// UUID, See man journalctl --new-id128
	sd_message_id = "MESSAGE_ID"
)

type remove_ansi_escape int

const (
	// bit flags
	Remove_journal remove_ansi_escape = 1 << iota
	Remove_writer
)

type Writer_option struct {
	Color        string
	Include_file bool
}

var (
	id128                      map[string]interface{}
	default_writer             io.Writer
	default_remove_ansi_escape remove_ansi_escape
	default_color              = map[Priority]Writer_option{
		Log_alert:   Writer_option{ansi.ColorCode("red+bh"), true},
		Log_crit:    Writer_option{ansi.ColorCode("red+bh"), true},
		Log_err:     Writer_option{ansi.ColorCode("red+bh"), true},
		Log_warning: Writer_option{ansi.ColorCode("208+bh"), true}, // orange
		Log_notice:  Writer_option{ansi.ColorCode("208+bh"), true}, // orange
		Log_info:    Writer_option{``, false},
	}
	default_disable_journal = false
	default_use_color       = true
	package_lock            sync.Mutex
	message_priority        = map[string]interface{}{Sd_message: ``, sd_priority: ``}
	valid_field             = regexp.MustCompile(`^[^_]{1}[\p{Lu}0-9_]*$`)
	max_fields              = uint64(C.sysconf(C._SC_IOV_MAX))
	sd_field_name_sep_s     = string(sd_field_name_sep_b)
	sd_field_name_sep_b     = []byte{61}
	remove_re2              = regexp.MustCompile(`\x1b[^m]*m`)
)

// See http://www.freedesktop.org/software/systemd/man/SD_JOURNAL_SUPPRESS_LOCATION.html,
// or man sd_journal_print, for valid systemd journal fields.
const (
	Sd_message = "MESSAGE"
	// Used in Set_default_fields(). systemd provides a default
	Sd_tag = "SYSLOG_IDENTIFIER"
)

// Journal can contain default systemd fields.
// See Set_default_fields().
type Journal struct {
	default_fields     map[string]interface{}
	lock               sync.Mutex
	add_go_code_fields bool
	writer             io.Writer
	stack_skip         int
	remove             remove_ansi_escape
	priority           Priority
}

type option func(o *Journal) option

func Set_remove_ansi(rm remove_ansi_escape) option {
	return func(o *Journal) option {
		prev := o.remove
		o.remove = rm
		return Set_remove_ansi(prev)
	}
}

// Sets the package level/default remove_ansi_escape and the current
// *Journal intance.
// Returns previous default remove_ansi_escape.
//
func Set_default_remove_ansi(rm remove_ansi_escape) option {
	return func(o *Journal) option {
		package_lock.Lock()
		defer package_lock.Unlock()
		prev := default_remove_ansi_escape
		default_remove_ansi_escape = rm
		o.remove = default_remove_ansi_escape
		return Set_default_remove_ansi(prev)
	}
}

// Sets the journal field name to value. The field will
// be removed when value is nil. An invalid name will be
// silently ignored. See info for Sd_tag.
//
func Set_field(name string, value interface{}) option {
	if valid_field.FindString(name) == "" {
		return func(o *Journal) option {
			return Set_field(``, nil)
		}
	}
	if value == nil {
		return func(o *Journal) option {
			prev := o.default_fields[name]
			delete(o.default_fields, name)
			return Set_field(name, prev)
		}
	} else {
		return func(o *Journal) option {
			prev := o.default_fields[name]
			o.default_fields[name] = value
			return Set_field(name, prev)
		}
	}
}

func Set_priority(p Priority) option {
	return func(o *Journal) option {
		prev := o.priority
		o.priority = p
		return Set_priority(prev)
	}
}

func Set_writer(w io.Writer) option {
	return func(o *Journal) option {
		prev := o.writer
		o.writer = w
		return Set_writer(prev)
	}
}

// New makes a Journal
//
func New(opt ...option) *Journal {
	r := New_journal_m(nil)
	r.Option(opt...)
	return r
}

// New_journal makes a Journal.
//
func New_journal() *Journal {
	return New_journal_m(nil)
}

// New_journal_m makes a Journal. The allowable interface{} values are
// string and []byte. A copy of []byte is made.
//
func New_journal_m(default_fields map[string]interface{}) *Journal {
	package_lock.Lock()
	j := &Journal{
		add_go_code_fields: true,
		priority:           Log_info,
		remove:             default_remove_ansi_escape,
		writer:             default_writer,
		stack_skip:         4,
	}
	package_lock.Unlock()
	j.Set_default_fields(default_fields)
	return j
}

// Option sets the options specified.
// It returns an option to restore the last arg's previous value.
//
func (o *Journal) Option(opt ...option) (previous option) {
	o.lock.Lock()
	defer o.lock.Unlock()
	for _, i := range opt {
		previous = i(o)
	}
	return
}

// Copy copies maps into a new map.
//
func (j *Journal) copy(maps ...map[string]interface{}) map[string]interface{} {
	j.lock.Lock()
	defer j.lock.Unlock()
	dest := make(map[string]interface{}, 3)
	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				switch t := v.(type) {
				case Priority:
					if 0 < len(string(t)) {
						dest[k] = v
					}
				case string:
					if 0 < len(string(t)) {
						dest[k] = v
					}
				case []byte:
					if 0 < len([]byte(t)) {
						dest[k] = append([]byte{}, t...)
					}
				}
			}
		}
	}
	return dest
}

// Default fields are sent with every Send().
// Do not include MESSAGE, or Priority, as these fields are always sent. The
// allowable interface{} values are string and []byte. A copy of []byte is
// made.
//
func (j *Journal) Set_default_fields(fields map[string]interface{}) {
	j.default_fields = j.copy([]map[string]interface{}{fields, message_priority, id128}...)
}

func (j *Journal) load_defaults(message string, Priority Priority) map[string]interface{} {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.default_fields[Sd_message] = message
	j.default_fields[sd_priority] = Priority
	if id128 == nil {
		delete(j.default_fields, sd_message_id)
	} else {
		j.default_fields[sd_message_id] = id128[sd_message_id]
	}
	return j.default_fields
}

// Set_writer_priority set the priority for the write() receiver.
// You'll probably want to use Set_remove_ansi(sd.Remove_journal).
// Default: Log_info.
//
func (j *Journal) Set_writer_priority(p Priority) *Journal {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.priority = p
	return j
}

// Writer implements io.Writer.
// Allows Journal to be used in the log package.
// You might want to use Set_remove_ansi(true).
// See http://godoc.org/log#SetOutput.
//
func (j *Journal) Write(b []byte) (int, error) {
	return len(b), j.Send(j.load_defaults(string(b), j.priority))
}

func (j *Journal) Emerg(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_emerg))
}

// Alert sends a message with Log_alert Priority (syslog severity).
// a ...interface{}: fmt.Println formating will become MESSAGE; see man
// systemd.journal-fields.
//
func (j *Journal) Alert(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_alert))
}

func (j *Journal) Crit(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_crit))
}

func (j *Journal) Err(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_err))
}

func (j *Journal) Warning(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_warning))
}

func (j *Journal) Notice(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_notice))
}

func (j *Journal) Info(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_info))
}

func (j *Journal) Debug(a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintln(a...), Log_debug))
}

func (j *Journal) Emerg_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_emerg)}...))
}

// Alert_m sends a message with Log_alert Priority (syslog severity).
// fields: your user-defined systemd.journal-fields.
// a ...interface{}: fmt.Println formating will become MESSAGE; see man
// systemd.journal-fields.
//
func (j *Journal) Alert_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_alert)}...))
}

func (j *Journal) Crit_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_crit)}...))
}

func (j *Journal) Err_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_err)}...))
}

func (j *Journal) Warning_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_warning)}...))
}

func (j *Journal) Notice_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_notice)}...))
}

func (j *Journal) Info_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_info)}...))
}

func (j *Journal) Debug_m(fields map[string]interface{}, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintln(a...), Log_debug)}...))
}

func (j *Journal) Emerg_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_emerg)}...))
}

// Alert_m_f sends a message with Log_alert Priority (syslog severity). The
// message is formed via fmt.Printf style arguments fields: your
// user-defined systemd.journal-fields. format string, a ...interface{}:
// see fmt.Printf.
//
func (j *Journal) Alert_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_alert)}...))
}

func (j *Journal) Crit_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_crit)}...))
}

func (j *Journal) Err_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_err)}...))
}

func (j *Journal) Warning_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_warning)}...))
}

func (j *Journal) Notice_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_notice)}...))
}

func (j *Journal) Info_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_info)}...))
}

func (j *Journal) Debug_m_f(fields map[string]interface{}, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{fields, j.load_defaults(fmt.Sprintf(format, a...), Log_debug)}...))
}

// Alertf sends a message with Log_alert Priority (syslog severity). The
// message is formed via fmt.Printf style arguments format string, a
// ...interface{}: see fmt.Printf.
//
func (j *Journal) Alertf(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_alert))
}

func (j *Journal) Critf(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_crit))
}

func (j *Journal) Errf(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_err))
}

func (j *Journal) Warningf(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_warning))
}

func (j *Journal) Noticef(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_notice))
}

func (j *Journal) Infof(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_info))
}

func (j *Journal) Debugf(format string, a ...interface{}) error {
	return j.Send(j.load_defaults(fmt.Sprintf(format, a...), Log_debug))
}

func (j *Journal) a_to_map(fields []string) (ret map[string]interface{}) {
	ret = make(map[string]interface{}, len(fields))
	for _, s := range fields {
		f := strings.SplitN(s, "=", 2)
		if len(f) == 2 {
			ret[f[0]] = f[1]
		}
	}
	return ret
}

// Alert_a sends a message with Log_alert Priority (syslog severity). fields:
// your user-defined systemd.journal-fields. a ...interface{}: fmt.Println
// formating will become MESSAGE; see man systemd.journal-fields.
//
func (j *Journal) Alert_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_alert)}...))
}

func (j *Journal) Crit_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_crit)}...))
}

func (j *Journal) Err_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_err)}...))
}

func (j *Journal) Warning_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_warning)}...))
}

func (j *Journal) Notice_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_notice)}...))
}

func (j *Journal) Info_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_info)}...))
}

func (j *Journal) Debug_a(fields []string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintln(a...), Log_debug)}...))
}

// Alert_a_f sends a message with Log_alert Priority (syslog severity). The
// message is formed via fmt.Printf style arguments fields: your
// user-defined systemd.journal-fields. format string, a ...interface{}:
// see fmt.Printf.
//
func (j *Journal) Alert_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_alert)}...))
}

func (j *Journal) Crit_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_crit)}...))
}

func (j *Journal) Err_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_err)}...))
}

func (j *Journal) Warning_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_warning)}...))
}

func (j *Journal) Notice_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_notice)}...))
}

func (j *Journal) Info_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_info)}...))
}

func (j *Journal) Debug_a_f(fields []string, format string, a ...interface{}) error {
	return j.Send(j.copy([]map[string]interface{}{j.a_to_map(fields), j.load_defaults(fmt.Sprintf(format, a...), Log_debug)}...))
}

// Set_add_go_code_fields will add GO_FILE (<file name>#<line #>),and GO_FUNC
// fields to the journal Send() methods, Info(), Err(), Warning(), etc..
// Default: use_go_code_fields = true.
//
func (j *Journal) Set_add_go_code_fields(use bool) {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.add_go_code_fields = use
}

// Useful when file/line are not correct
// default: 4
func (j *Journal) Stack_skip(skip int) *Journal {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.stack_skip = skip
	return j
}

// Set_message_id sets the systemd MESSAGE_ID (UUID) for all Journal
// (Global) instances. Generate an application UUID with journalctl
// --new-id128. See man journalctl.
//
// uuid is unset with ""
//
func Set_message_id(uuid string) {
	package_lock.Lock()
	defer package_lock.Unlock()
	if uuid == "" {
		id128 = nil
	} else {
		id128 = map[string]interface{}{sd_message_id: uuid}
	}
}

func Set_default_writer_stderr() option {
	return Set_default_writer(os.Stderr)
}

func Set_default_writer_stdout() option {
	return Set_default_writer(os.Stdout)
}

// Set output to an additional io.Writer
//
func Set_default_writer(w io.Writer) option {
	return func(o *Journal) option {
		package_lock.Lock()
		defer package_lock.Unlock()
		prev := default_writer
		default_writer = w
		return Set_default_writer(prev)
	}
}

// Set default colors for io.Writer.
//
// default: red (bold, highlight): Log_alert, Log_crti, Log_err, orange (bold, highlight):
// Log_warning, Log_notice
//
// example: map[Priority]string{Log_err: ansi.ColorCode("green")}
//
func Set_default_colors(colors map[Priority]Writer_option) {
	package_lock.Lock()
	defer package_lock.Unlock()
	default_color = colors
}

// Set default_remove_ansi_escape will set the default value for a new Journal.
//
func Set_default_remove_ansi_escape(rm remove_ansi_escape) {
	package_lock.Lock()
	defer package_lock.Unlock()
	default_remove_ansi_escape = rm
}

// Journal output will be disabled. Useful for just stdout/stderr logging with
// color.
//
func Set_default_disable_journal(disable bool) option {
	return func(o *Journal) option {
		package_lock.Lock()
		defer package_lock.Unlock()
		prev := default_disable_journal
		default_disable_journal = disable
		return Set_default_disable_journal(prev)
	}
}

// Send writes to the systemd-journal. The keys must be uppercase strings
// without a leading _. The other send methods are easier to use. See Info(),
// Infom(), Info_m_f(), etc. A MESSAGE key in field is the only required
// field.
//
func (j *Journal) Send(fields map[string]interface{}) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	package_lock.Lock()
	disable_journal := default_disable_journal
	package_lock.Unlock()
	w := j.writer
	if w == nil {
		package_lock.Lock()
		w = default_writer
		package_lock.Unlock()
	}
	if s, ok := fields[Sd_message].(string); ok {
		var priority Priority
		if p, ok := fields[sd_priority].(Priority); ok {
			priority = Priority(p)
		}
		var cleaned_s string
		// writer
		if w != nil {
			if j.remove&Remove_writer != 0 {
				cleaned_s = remove_re2.ReplaceAllLiteralString(s, ``)
				if default_use_color {
					package_lock.Lock()
					var line string
					if default_color[priority].Include_file {
						if j.add_go_code_fields {
							_, f, l := file_line(j.stack_skip)
							line = fmt.Sprintf("%v:%v ", f, l)
						}
					}
					reset := ``
					if 0 < len(default_color[priority].Color) {
						reset = ansi.Reset
					}
					fmt.Fprintf(w, "%v%v%v%v", default_color[priority].Color, line, cleaned_s, reset)
					package_lock.Unlock()
				} else {
					fmt.Fprintf(w, cleaned_s)
				}
			} else {
				if default_use_color {
					package_lock.Lock()
					var line string
					if default_color[priority].Include_file {
						if j.add_go_code_fields {
							_, f, l := file_line(j.stack_skip)
							line = fmt.Sprintf("%v:%v ", f, l)
						}
					}
					reset := ``
					if 0 < len(default_color[priority].Color) {
						reset = ansi.Reset
					}
					fmt.Fprintf(w, "%v%v%v%v", default_color[priority].Color, line, s, reset)
					package_lock.Unlock()
				} else {
					fmt.Fprintf(w, s)
				}
			}
		}
		if disable_journal {
			return nil
		}
		// journal
		if j.remove&Remove_journal != 0 {
			if 0 == len(cleaned_s) {
				fields[Sd_message] = remove_re2.ReplaceAllLiteralString(s, ``)
			} else {
				fields[Sd_message] = cleaned_s
			}
		}
	}
	// journal
	if max_fields < uint64(len(fields)) {
		return errors.New(fmt.Sprintf("Field count cannot exceed %v: %v given", max_fields, len(fields)))
	}
	if j.add_go_code_fields {
		fn, file, line := file_line(j.stack_skip)
		fields[sd_go_func] = fn
		fields[sd_go_file] = file + `:` + strconv.Itoa(line)
	}
	iov := C.malloc(C.size_t(C.sizeof_struct_iovec * len(fields)))
	i := 0
	defer func() {
		for j := 0; j < i; j++ {
			C.free(((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(j)*C.sizeof_struct_iovec))).iov_base)
		}
		C.free(iov)
	}()
	for k, v := range fields {
		if valid_field.FindString(k) == "" {
			return fmt.Errorf("field violates regexp %v : %v", valid_field, k)
		}
		switch t := v.(type) {
		case string:
			s := k + sd_field_name_sep_s + t
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_base = unsafe.Pointer(C.CString(s))
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_len = C.size_t(len(s))
		case Priority:
			s := k + sd_field_name_sep_s + string(t)
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_base = unsafe.Pointer(C.CString(s))
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_len = C.size_t(len(s))
		case []byte:
			b := bytes.Join([][]byte{[]byte(k), t}, sd_field_name_sep_b)
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_base = C.CBytes(b)
			((*C.struct_iovec)(unsafe.Pointer(uintptr(iov) + uintptr(i)*C.sizeof_struct_iovec))).iov_len = C.size_t(len(b))
		default:
			return fmt.Errorf("Error: Unsupported field value: key = %v", k)
		}
		i++
	}
	n, _ := C.sd_journal_sendv((*C.struct_iovec)(iov), C.int(len(fields)))
	if n != 0 {
		return errors.New("Error with sd_journal_sendv arguments")
	}
	return nil
}

// 4
func file_line(skip int) (fn string, file string, line int) {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return ``, ``, 0
	}
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function, trim_go_path(frame.Function, frame.File), frame.Line
}

func trim_go_path(name, file string) string {
	// From github.com/pkg/errors, BSD-2-Clause
	const sep = "/"
	goal := strings.Count(name, sep) + 2
	i := len(file)
	for n := 0; n < goal; n++ {
		i = strings.LastIndex(file[:i], sep)
		if i == -1 {
			i = -len(sep)
			break
		}
	}
	file = file[i+len(sep):]
	return file
}
