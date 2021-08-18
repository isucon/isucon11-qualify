require_relative './app'

use Rack::Logger
use Rack::CommonLogger
run Isucondition::App

