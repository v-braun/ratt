vars {
  baseUrl = "https://62508d0de3e5d24b342440f9.mockapi.io"
  newUserId = ""
  roland = {
      firstName: "Roland",
      lastName: "Deschain"
    }
}

request "create_user" {  
  args{
    baseUrl = ""
    firstName = ""
    lastName = ""
  }

  url = "${vars.baseUrl}/user"
  method = "POST"
  body = vars.roland
}

request "get_user" {  
  args{
    userId = ""
  }

  url = "${vars.baseUrl}/user/${args.userId}"
  method = "GET"
}

script "example_script" {
  content = <<EOT

    sleep(2000)

EOT
}



