# Predict Search

*Note: register for an Clarifai API key first*

*Build the app*
Copy project folder to /usr/local/src then build the app using
`go build predict`

*Environment Variable*
Set the environment variable `PREDICT_API_KEY` with your own Clarifai API key.

*Run the app*
Firstly build the tag-mapping json file with:
`go run predict -build [Default=images.txt | path to the image-url .txt file]`
Which will generate a TagMap.json in the project folder.

When it's built, open
`http://localhost:5000`

Type into the search bar for a keyword. It will respond with the top-10 most relative image from the image-url file.

When TagMap.json is already in the folder, run the app with:
`go run predict`

Rebuild the TagMap.json when the source images need to be changed, with the flag `-build` following with the path to the image-url file.