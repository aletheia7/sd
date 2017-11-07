// Copyright 2016 aletheia7. All rights reserved. Use of this source code is
// governed by a BSD-2-Clause license that can be found in the LICENSE file.
/*

Package sd provides methods to write to the systemd-journal.

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

package sd

import (
	"fmt"
	"log/syslog"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	Sd_message            = "MESSAGE"
	sd_valid_field_regexp = `^[^_]{1}[\p{Lu}0-9_]*$`
	sd_go_func            = "GO_FUNC"
	sd_go_file            = "GO_FILE"
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

type Send_stderr int

const (
	Sd_send_stderr_allow_override Send_stderr = iota
	Sd_send_stderr_true                       = iota
	Sd_send_stderr_false                      = iota
)

var (
	id128                      map[string]interface{}
	default_send_stderr        = Sd_send_stderr_allow_override
	default_remove_ansi_escape = false
	package_lock               sync.Mutex
	message_priority           = map[string]interface{}{Sd_message: ``, sd_priority: ``}
	sd_priority                = "PRIORITY"
)

// See http://www.freedesktop.org/software/systemd/man/SD_JOURNAL_SUPPRESS_LOCATION.html,
// or man sd_journal_print, for valid systemd journal fields.
const (
	// Used in Set_default_fields(). systemd provides a default
	Sd_tag = "SYSLOG_IDENTIFIER"
	// UUID, See man journalctl --new-id128
	sd_message_id = "MESSAGE_ID"
)

// Journal can contain default systemd fields.
// See Set_default_fields().
type Journal struct {
	default_fields     map[string]interface{}
	lock               sync.Mutex
	add_go_code_fields bool
	send_stderr        Send_stderr
	remove_ansi_escape bool
	writer_priority    Priority
	remove_re2         *regexp.Regexp
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
	j := &Journal{
		add_go_code_fields: true,
		remove_ansi_escape: default_remove_ansi_escape,
		writer_priority:    Log_info,
	}
	j.remove_re2 = regexp.MustCompile(`\x1b[^m]*m`)
	j.Set_default_fields(default_fields)
	return j
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
	switch j.remove_ansi_escape {
	case true:
		j.default_fields[Sd_message] = j.remove_re2.ReplaceAllLiteralString(message, ``)
	case false:
		j.default_fields[Sd_message] = message
	}
	j.default_fields[sd_priority] = Priority

	if id128 == nil {
		delete(j.default_fields, sd_message_id)
	} else {
		j.default_fields[sd_message_id] = id128[sd_message_id]
	}
	return j.default_fields
}

// Set_remove_ansi_escape determines if ANSI escape sequences are removed.
// Default: false.
//
func (j *Journal) Set_remove_ansi_escape(remove bool) {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.remove_ansi_escape = remove
}

// Set_writer_priority set the priority for the write() receiver.
// You'll probably want to use Set_remove_ansi(true).
// Default: Log_info.
//
func (j *Journal) Set_writer_priority(p Priority) {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.writer_priority = p
}

// Writer implements io.Writer.
// Allows Jhournal to be used in the log package.
// You might want to use Set_remove_ansi(true).
// See http://godoc.org/log#SetOutput.
//
func (j *Journal) Write(b []byte) (int, error) {
	return len(b), j.Send(j.load_defaults(string(b), j.writer_priority))
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

// Set_send_stderr to Sd_send_stderr_true to send a message to os.Stderr in
// addition to the systemd journal. Set_send_stderr to Sd_send_stderr_false
// to prevent sending to os.Stderr. This will override
// Set_default_stderr_override() Default: Sd_send_stderr_override; i.e.
// allow Set_default_stderr_override() the first choice.
//
func (j *Journal) Set_send_stderr(use Send_stderr) {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.send_stderr = use
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

// Set_default_send_stderr to Sd_send_stderr_true to send a message to
// os.Stderr in addition to the journal. Can be overridden when
// Journal.Set_send_stderr(Sd_send_stderr_true) is called. Default:
// Sd_send_stderr_override; i.e. will not send to stderr.
//
func Set_default_send_stderr(use Send_stderr) {
	package_lock.Lock()
	defer package_lock.Unlock()
	default_send_stderr = use
}

// Set default_remove_ansi_escape will set the default value for new Journal.
// Default: remove = false.
//
func Set_default_remove_ansi_escape(remove bool) {
	package_lock.Lock()
	defer package_lock.Unlock()
	default_remove_ansi_escape = remove
}
