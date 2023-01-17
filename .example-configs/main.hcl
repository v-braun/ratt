
testcase "create_and_get_user" {
  invoke "request" "create_user"{
    firstName = "Paul"
    lastName = "Panzer"

    then "assert" {
      status_to_be_201 = (res.statusCode == 201)
      server_is_not_nginx = (res.headers["Server"] != "nginx")
    }
    then "set" {
      newUserId = res.body.id
    }
    then "assert" {
      status_to_be_200 = (res.statusCode == 201)
    }
  }

  invoke script "example_script"{
    
  }

  invoke "request" "get_user"{
    userId = "${vars.newUserId}"

    then "assert" {
      status_to_be_200 = (res.statusCode == 200)
      user_name_is_roland = (res.body.firstName == "Roland")
    }
  }
}

