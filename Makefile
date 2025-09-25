# Build all target
.PHONY : all
all : app

# Link the object files and dependent libraries into a binary
app : VMTranslator
	echo
