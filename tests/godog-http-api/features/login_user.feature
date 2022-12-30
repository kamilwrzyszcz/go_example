Feature: Tests related to login of a user

  @success
  Scenario: Successfully create user and login

    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_NAME"
    Given I generate a random word having from "6" to "10" of "english" characters and save it as "RANDOM_PASSWORD"
    Given I save "{{.RANDOM_NAME}}@gmail.com" as "RANDOM_EMAIL"

    When I send "POST" request to "{{.MY_APP_URL}}/users" with body and headers:
    """
    {
      "body": {
        "email": "{{.RANDOM_EMAIL}}",
        "full_name": "{{.RANDOM_NAME}} Doe",
        "password": "{{.RANDOM_PASSWORD}}",
        "username": "{{.RANDOM_NAME}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 201

    When I send "POST" request to "{{.MY_APP_URL}}/users/login" with body and headers:
    """
    {
      "body": {
        "password": "{{.RANDOM_PASSWORD}}",
        "username": "{{.RANDOM_NAME}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should be 200
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" response should have nodes "session_id, access_token, access_token_expires_at, refresh_token, refresh_token_expires_at, user"
    And the "JSON" node "user.username" should be "string" of value "{{.RANDOM_NAME}}"
    And the "JSON" node "user.full_name" should be "string" of value "{{.RANDOM_NAME}} Doe"
    And the "JSON" node "user.email" should be "string" of value "{{.RANDOM_EMAIL}}"
    And the "JSON" response should have nodes "user.created_at, user.password_changed_at"

  @failure
  Scenario: Unsuccessful attempt to login a user due to invalid user

    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_NAME"
    Given I generate a random word having from "6" to "10" of "english" characters and save it as "RANDOM_PASSWORD"
    Given I save "{{.RANDOM_NAME}}@gmail.com" as "RANDOM_EMAIL"

    When I send "POST" request to "{{.MY_APP_URL}}/users" with body and headers:
    """
    {
      "body": {
        "email": "{{.RANDOM_EMAIL}}",
        "full_name": "{{.RANDOM_NAME}} Doe",
        "password": "{{.RANDOM_PASSWORD}}",
        "username": "{{.RANDOM_NAME}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 201

    When I send "POST" request to "{{.MY_APP_URL}}/users/login" with body and headers:
    """
    {
      "body": {
        "password": "{{.RANDOM_PASSWORD}}",
        "username": "{{.RANDOM_NAME}}XYZ"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 404
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" node "error" should be "string"

  @failure
  Scenario: Unsuccessful attempt to login a user due to invalid password

    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_NAME"
    Given I generate a random word having from "6" to "10" of "english" characters and save it as "RANDOM_PASSWORD"
    Given I save "{{.RANDOM_NAME}}@gmail.com" as "RANDOM_EMAIL"

    When I send "POST" request to "{{.MY_APP_URL}}/users" with body and headers:
    """
    {
      "body": {
        "email": "{{.RANDOM_EMAIL}}",
        "full_name": "{{.RANDOM_NAME}} Doe",
        "password": "{{.RANDOM_PASSWORD}}",
        "username": "{{.RANDOM_NAME}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 201

    When I send "POST" request to "{{.MY_APP_URL}}/users/login" with body and headers:
    """
    {
      "body": {
        "password": "{{.RANDOM_PASSWORD}}XYZ",
        "username": "{{.RANDOM_NAME}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 401
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" node "error" should be "string"

  @failure
  Scenario: Unsuccessful attempt to login a user due to malformed request

    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_NAME"
    Given I generate a random word having from "6" to "10" of "english" characters and save it as "RANDOM_PASSWORD"
    Given I save "{{.RANDOM_NAME}}@gmail.com" as "RANDOM_EMAIL"

    When I send "POST" request to "{{.MY_APP_URL}}/users" with body and headers:
    """
    {
      "body": {
        "email": "{{.RANDOM_EMAIL}}",
        "full_name": "{{.RANDOM_NAME}} Doe",
        "password": "{{.RANDOM_PASSWORD}}",
        "username": "{{.RANDOM_NAME}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 201

    When I send "POST" request to "{{.MY_APP_URL}}/users/login" with body and headers:
    """
    {
      "body": {
        "password": "{{.RANDOM_PASSWORD}}"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 200
    But the response status code should be 400
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" node "error" should be "string"