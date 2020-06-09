window.onload = function() {

  var x = document.cookie;
  this.console.log(x);
  if (x == "") {
    var tmp = document.getElementById('login')

    tmp.setAttribute('href', '/login')
    tmp = document.getElementById('register')
    tmp.setAttribute('href', '/signup')
    tmp.innerHTML = "Register"
  } else {
    var tmp = document.getElementById('login')
    this.console.log("here")
    tmp.setAttribute('href', '/profile')
    tmp = document.getElementById('register')
    tmp.setAttribute('href', '/logout')
    tmp.innerHTML = "logout"
  }

}


// function change() {
//   document.getElementById('gandu').innerHTML = "hello";
// }