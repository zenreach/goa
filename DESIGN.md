goa.MediaType:
  - provides Render() used by goa that produces JSON, can be overridden 
  - Actions only need to return values for responses that have bodies
  - One return value per different response with body