Feature: Tests related to creation of user

  @success
  Scenario: Successfully create user

    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_NAME"
    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_PASSWORD"
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
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" node "email" should be "string" of value "{{.RANDOM_EMAIL}}"
    And the "JSON" node "full_name" should be "string" of value "{{.RANDOM_NAME}} Doe"
    And the "JSON" node "username" should be "string" of value "{{.RANDOM_NAME}}"
    And the "JSON" response should have nodes "created_at, password_changed_at"

  @failure
  Scenario: Unsuccessful attempt to create user due to invalid request body
    When I send "POST" request to "{{.MY_APP_URL}}/users" with body and headers:
    """
    {
      "body": {
        "email": "a@b.c"
      },
      "headers": {
        "Content-Type": "application/json"
      }
    }
    """
    Then the response status code should not be 201
    But the response status code should be 400
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" node "error" should be "string"

  @failure
  Scenario: Unsuccessful attempt to create use due to email duplication

    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_NAME"
    Given I generate a random word having from "5" to "10" of "english" characters and save it as "RANDOM_PASSWORD"
    Given I save "{{.RANDOM_NAME}}@gmail.com" as "RANDOM_EMAIL"

    # Successfully create random user with email equal to {{.RANDOM_EMAIL}}
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
    Then the response status code should be 201

    # Unsuccessful attempt to create user with the same email equal to {{.RANDOM_EMAIL}}
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
    Then the response status code should be 403
    And the response should have header "Content-Type" of value "application/json; charset=utf-8"
    And the response body should have format "JSON"
    And the "JSON" node "error" should be "string"
    And the "JSON" node "error" should contain sub string "duplicate key"