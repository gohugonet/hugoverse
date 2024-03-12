package admin

func (v *View) Contents(t string, status string) string {
	return `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="col s8">
						<div class="row">
							<div class="card-title col s7">` + t + ` Items</div>
							<div class="col s5 input-field inline">
								<select class="browser-default __ponzu sort-order">
									<option value="DESC">New to Old</option>
									<option value="ASC">Old to New</option>
								</select>
								<label class="active">Sort:</label>
							</div>	
							<script>
								$(function() {
									var sort = $('select.__ponzu.sort-order');

									sort.on('change', function() {
										var path = window.location.pathname;
										var s = sort.val();
										var t = getParam('type');
										var status = getParam('status');

										if (status == "") {
											status = "public";
										}

										window.location.replace(path + '?type=' + t + '&order=' + s + '&status=' + status);
									});

									var order = getParam('order');
									if (order !== '') {
										sort.val(order);
									}
									
								});
							</script>
						</div>
					</div>
					<form class="col s4" action="/admin/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                    </form>	
					</div>`
}
