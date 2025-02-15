package admin

import (
	"bytes"
	"html/template"
)

// Login ...
func Login(name string) ([]byte, error) {
	html := startAdminHTML + loginAdminHTML + endAdminHTML

	a := View{
		Logo: name,
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("login").Parse(html))
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (v *View) UserManagementView(data map[string]interface{}) (_ []byte, err error) {
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("users").Parse(v.UserManagement()))
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	return v.SubView(buf.Bytes())
}

func (v *View) UserManagement() string {
	html := `
    <div class="card user-management">
        <div class="card-title">Edit your account:</div>    
        <form class="row" enctype="multipart/form-data" action="/admin/configure/users/edit" method="post">
            <div class="col s9">
                <label class="active">Email Address</label>
                <input type="email" name="email" value="{{ .User.Name }}"/>
            </div>

            <div class="col s9">
                <div>To approve changes, enter your password:</div>
                
                <label class="active">Current Password</label>
                <input type="password" name="password"/>
            </div>

            <div class="col s9">
                <label class="active">New Password: (leave blank if no password change needed)</label>
                <input name="new_password" type="password"/>
            </div>

            <div class="col s9">                        
                <button class="btn waves-effect waves-light green right" type="submit">Save</button>
            </div>
        </form>

        <div class="card-title">Add a new user:</div>        
        <form class="row" enctype="multipart/form-data" action="/admin/configure/users" method="post">
            <div class="col s9">
                <label class="active">Email Address</label>
                <input type="email" name="email" value=""/>
            </div>

            <div class="col s9">
                <label class="active">Password</label>
                <input type="password" name="password"/>
            </div>

            <div class="col s9">            
                <button class="btn waves-effect waves-light green right" type="submit">Add User</button>
            </div>   
        </form>        

        <div class="card-title">Remove Admin Users</div>        
        <ul class="users row">
            {{ range .Users }}
            <li class="col s9">
                {{ .Name }}
                <form enctype="multipart/form-data" class="delete-user __ponzu right" action="/admin/configure/users/delete" method="post">
                    <span>Delete</span>
                    <input type="hidden" name="email" value="{{ .Name }}"/>
                    <input type="hidden" name="id" value="{{ .ID }}"/>
                </form>
            </li>
            {{ end }}
        </ul>
    </div>
    `
	script := `
    <script>
        $(function() {
            var del = $('.delete-user.__ponzu span');
            del.on('click', function(e) {
                if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this user?\nThis cannot be undone.")) {
                    $(e.target).parent().submit();
                }
            });
        });
    </script>
    `

	return html + script
}
