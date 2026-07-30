# Columbia MS CS Course Advisor RAG

An intelligent, AI-powered course advisory assistant for Columbia 
University students. Built using **Retrieval-Augmented Generation (RAG)**, 
this application allows students to query course offerings, prerequisites, abd syllabus details using natural language.

---

## Features

- **Natural Language Course Search:** Ask questions like *"What machine learning courses can I take if I've only taken Intro to CS?"* or *"Which computer science electives focus heavily on group projects?"*
- **Context-Aware Retrieval:** Uses vector embeddings to retrieve relevant details directly from official Columbia course catalogs, bulletin pages, and syllabi.
- **Accurate & Grounded Answers:** Leverages LLMs with grounded retrieval to minimize hallucinations and provide reliable academic guidance.
- **Prerequisite & Track Analysis:** Easily verify course prerequisites, core requirements, and degree track alignment.

---

## Architecture

```text
┌─────────────────┐       ┌────────────────┐       ┌─────────────────┐
│ User Query      │ ───>  │ Embedding Model│ ───>  │ Vector DB       │
└─────────────────┘       └────────────────┘       │                 │
                                                   └────────┬────────┘
                                                            │ Relevant Chunks
                                                            v
┌─────────────────┐       ┌────────────────┐       ┌─────────────────┐
│ Final Answer    │ <───  │ LLM            │ <───  │ Prompt +        │
└─────────────────┘       └────────────────┘       │ Context         │
                                                   └─────────────────┘
```
## 🌟 API Example
```code

cURL

postman request POST 'http://localhost:8080/recommend' \
  --header 'Content-Type: application/json' \
  --body '{
    "interests": "I am interested in deep learning and neural networks",
    "completed_courses": ["COMS 4771", "COMS 6111", "COMS 4701"],
    "gpa": 3.5,
    "pathway": "Machine Learning"
  }'
  
  
sample response
{
    "recommendations": [
        {
            "CourseCode": "ECBM E4040",
            "Content": "Developing features - internal representations of the world, artificial neural networks, classifying handwritten digits with logistics regression, feedforward deep networks, back propagation in multilayer perceptrons, regularization of deep or distributed models, optimization for training deep models, convolutional neural networks, recurrent and recursive neural networks, deep learning in speech and object recognition",
            "Distance": 0.40253034815011224,
            "SatisfiesPathway": false,
            "PathwayGroup": "",
            "PathwayTitle": "",
            "SatisfiesBreadth": false,
            "BreadthCategory": ""
        },
        {
            "CourseCode": "COMS W4732",
            "Content": "Advanced course in computer vision. Topics include convolutional networks and back-propagation, object and action recognition, self-supervised and few-shot learning, image synthesis and generative models, object tracking, vision and language, vision and audio, 3D representations, interpretability, and bias, ethics, and media deception",
            "Distance": 0.5447087857300759,
            "SatisfiesPathway": false,
            "PathwayGroup": "",
            "PathwayTitle": "",
            "SatisfiesBreadth": false,
            "BreadthCategory": ""
        },
        {
            "CourseCode": "EECS E6898",
            "Content": "Advanced topics spanning electrical engineering and computer science such as speech processing and recognition, image and multimedia content analysis, and other areas drawing on signal processing, information theory, machine learning, pattern recognition, and related topics. Content varies from year to year, and different topics rotate through the course numbers 6890 to 6899",
            "Distance": 0.5679536800356844,
            "SatisfiesPathway": false,
            "PathwayGroup": "",
            "PathwayTitle": "",
            "SatisfiesBreadth": false,
            "BreadthCategory": ""
        },
        {
            "CourseCode": "EECS E6897",
            "Content": "Advanced topics spanning electrical engineering and computer science such as speech processing and recognition, image and multimedia content analysis, and other areas drawing on signal processing, information theory, machine learning, pattern recognition, and related topics. Content varies from year to year, and different topics rotate through the course numbers 6890 to 6899",
            "Distance": 0.5679536800356844,
            "SatisfiesPathway": false,
            "PathwayGroup": "",
            "PathwayTitle": "",
            "SatisfiesBreadth": false,
            "BreadthCategory": ""
        },
        {
            "CourseCode": "EECS E6894",
            "Content": "Advanced topics spanning electrical engineering and computer science such as speech processing and recognition, image and multimedia content analysis, and other areas drawing on signal processing, information theory, machine learning, pattern recognition, and related topics. Content varies from year to year, and different topics rotate through the course numbers 6890 to 6899",
            "Distance": 0.5679536800356844,
            "SatisfiesPathway": false,
            "PathwayGroup": "",
            "PathwayTitle": "",
            "SatisfiesBreadth": false,
            "BreadthCategory": ""
        }
    ],
    "explanation": "Based on your interest in deep learning and neural networks, here are the recommendations from your candidate list:\n\n*   **ECBM E4040**: This course directly aligns with your goals, covering artificial neural networks, deep networks, back propagation, and deep learning applications. It is a general elective and does not fill a specific pathway or breadth slot.\n*   **COMS W4732**: This course covers advanced computer vision topics, including convolutional networks, generative models, and self-supervised learning. It is a general elective and does not satisfy any specific pathway or breadth requirement.\n*   **EECS E6898**: This course covers advanced topics in electrical engineering and computer science, such as machine learning and pattern recognition. It acts as a general elective and does not fill a specific pathway or breadth slot.\n*   **EECS E6897**: This course covers advanced topics spanning electrical engineering and computer science, including machine learning, signal processing, and pattern recognition. It is a general elective and does not satisfy a specific pathway or breadth requirement.\n*   **EECS E6894**: This course explores advanced topics like speech processing, machine learning, and pattern recognition. It is a general elective and does not fill any specific pathway or breadth slot."
} 
```
## .env format
```code
PGHOST=localhost
PGPORT=5432
PGDATABASE=course_advisor
PGUSER=myuser
PGPASSWORD=mypassword

# "anthropic" or "gemini" -- see generate/generate.go
LLM_PROVIDER=anthropic

ANTHROPIC_API_KEY=
ANTHROPIC_MODEL=claude-sonnet-4-6

GEMINI_API_KEY=<key>
GEMINI_MODEL=gemini-3.5-flash

EMBEDDING_SERVICE_URL=http://localhost:8001
```

## Running the services
- Step 1: Use scrapper to generate the raw HTMLs
- Step 2: Use the parser to generate clean texts from these HTMLs
- Step 3: Run the run load.py files in ingestion for both relational and vectors
- Step 4 : Create .env files at the root with the given format and add your keys
- Step 5: Spin up the services by  ``docker-compose up --build`` and then use the above curl