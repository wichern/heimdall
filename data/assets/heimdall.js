function set_script_running(id, running) {
  tr = $('#script_' + id)
  tr.removeClass('table-success')

  if (running) {
    tr.addClass('table-success')
  }
}

$(document).ready(function() {
  /*
   * On ready we fetch the list of known scripts.
   */
  $.ajax({
    url: 'api/scripts',
    method: 'GET',
  }).done(function(scripts) {
    for (script of scripts) {
      lastRunButton = ''
      if (script.Last != -1) {
        lastRunButton = `<button id="`
      }

      $('#scripts').append(`
        <tr id="script_${script.ID}" class="${script.Running ? 'table-success' : ''}">
          <td>${script.Name}</td>
          <td><a id="button_status" class="btn" href="status/${script.ID}/${script.Last}">Last</a></td>
          <td align="right">
              <div class="btn-group" role="group" aria-label="Actions">
                  <button 
                    id="button_start"
                    class="btn ml-2" 
                    onClick="if (!this.classList.contains('disabled')) { start_script(${script.ID}) }">
                      <i class="fa fa-play"></i>
                  </button>
                  <button 
                    id="button_stop" 
                    class="btn"
                    onClick="if (!this.classList.contains('disabled')) { stop_script(${script.ID}) }">
                      <i class="fa fa-stop"></i>
                  </button>
              </div>
          </td>
      </tr>`)

      enable_start_button(script.ID, !script.Running)
      enable_stop_button(script.ID, script.Running)
      enable_status_button(script.ID, script.Last != -1)
      set_script_running(script.ID, script.Running)
    }

    /*
     * Constantly update all script status.
     */
    (function pollStatus() {
      $.ajax({
        url: 'api/scripts',
        method: 'GET',
      }).done(function(scripts) {
        console.log(scripts)
        for (script of scripts) {
          enable_start_button(script.ID, !script.Running)
          enable_stop_button(script.ID, script.Running)

          tr = $('#script_' + script.ID)
          btn_status = tr.find('#button_status')
          btn_status.attr('href', 'status/' + script.ID + '/' + script.Last)
          enable_status_button(script.ID, script.Last != -1)
          set_script_running(script.ID, script.Running)
        }

        setTimeout(pollStatus, 1000);
      }).fail(function(jqXHR) {
        show_error(`Failed to get scripts status: ${jqXHR.responseText}`);
        setTimeout(pollStatus, 1000);
      });
    }());
  }).fail(function(jqXHR) {
    show_error(`Failed to get scripts: ${jqXHR.responseText}`);
  });
});

function enable_start_button(id, enable) {
  tr = $('#script_' + id)
  btn = tr.find('#button_start')

  if (enable) {
    btn.addClass('btn-success')
    btn.removeClass('disabled')
  } else {
    btn.removeClass('btn-success')
    btn.addClass('disabled')
  }

  btn.blur()
}

function enable_stop_button(id, enable) {
  tr = $('#script_' + id)
  btn = tr.find('#button_stop')

  if (enable) {
    btn.addClass('btn-danger')
    btn.removeClass('disabled')
  } else {
    btn.removeClass('btn-danger')
    btn.addClass('disabled')
  }

  btn.blur()
}

function enable_status_button(id, enable) {
  tr = $('#script_' + id)
  btn = tr.find('#button_status')

  if (enable) {
    btn.removeClass('disabled btn-outline-light')
    btn.addClass('btn-outline-primary')
  } else {
    btn.addClass('disabled btn-outline-light')
    btn.removeClass('btn-outline-primary')
  }

  btn.blur()
}

function start_script(id) {
  enable_start_button(id, false);

  $.ajax({
    url: 'api/scripts/start/' + id,
    method: 'POST',
  }).done(function(script) {
    console.log('started ' + id)
  }).fail(function(jqXHR) {
    show_error(`Failed to start script: ${jqXHR.responseText}`);
  });
}

function stop_script(id) {
  enable_stop_button(id, false);

  $.ajax({
    url: 'api/scripts/stop/' + id,
    method: 'POST',
  }).done(function(script) {
    console.log('stopped ' + id)
  }).fail(function(jqXHR) {
    show_error(`Failed to stop script: ${jqXHR.responseText}`);
  });
}

function show_error(msg) {
  errors = $('#errors');
  errors.empty();
  errors.append(`<div class="alert alert-danger" role="alert">${msg}</div>`);
}