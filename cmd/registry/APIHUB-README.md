#Registry tool

To run the binary on MacOS, you will need to unquarantine the binary.
> xattr -d com.apple.quarantine registry

###How to use registry command with Apigee API hub instance 

Steps:
1. Make sure you have gcloud command installed.
2. Setup the GCP_PROJECT variable.
3. Set you API hub project
    > gcloud config set project $GCP_PROJECT
4. Generate the token and set it in the environment variable using 
    > source auth/HOSTED.sh
  
   >For windows execute the `auth/HOSTED.bat` file
5. To list all the APIs in your API hub instance run the below command:
    > registry list projects/$GCP_PROJECT/locations/global/apis/-

6. Find a list of other supported commands  
    > registry help
